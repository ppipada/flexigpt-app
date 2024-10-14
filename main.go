//go:build !codeanalysis

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"context"

	"path/filepath"
	"strings"

	"github.com/flexigpt/flexiui/pkg/aiprovider"
	aiproviderSpec "github.com/flexigpt/flexiui/pkg/aiprovider/spec"
	"github.com/flexigpt/flexiui/pkg/conversationstore"
	"github.com/flexigpt/flexiui/pkg/settingstore"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/adrg/xdg"
)

const AppTitle = "FlexiGptUI"
const FRONTEND_PATH_PREFIX = "/frontend/build"

var DIR_PAGES = []string{"/agents", "/chats", "/settings", "/404"}

//go:embed all:packages/frontend/build
var assets embed.FS

//go:embed packages/frontend/public/icon.png
var icon []byte

// App struct
type App struct {
	ctx                  context.Context
	settingStoreAPI      *settingstore.SettingStore
	conversationStoreAPI *conversationstore.ConversationCollection
	providerSetAPI       *WrappedProviderSetAPI
	configBasePath       string
	dataBasePath         string
}

// NewApp creates a new App application struct
func NewApp() *App {
	if xdg.ConfigHome == "" || xdg.DataHome == "" {
		log.Panicf(
			"Could not resolve data paths. XDG Config dir: %s, XDG Data dir: %s",
			xdg.ConfigHome,
			xdg.DataHome,
		)
		return nil
	}

	app := &App{}
	app.configBasePath = filepath.Join(xdg.ConfigHome, (strings.ToLower(AppTitle)))
	app.dataBasePath = filepath.Join(xdg.DataHome, (strings.ToLower(AppTitle)))
	app.settingStoreAPI = &settingstore.SettingStore{}
	app.conversationStoreAPI = &conversationstore.ConversationCollection{}
	app.providerSetAPI = NewWrappedProviderSetAPI(aiproviderSpec.ProviderNameOpenAI)

	if err := os.MkdirAll(app.configBasePath, os.FileMode(0770)); err != nil {
		log.Panicf("Failed to create directories for config data %s: %v", app.configBasePath, err)
		return nil

	}
	if err := os.MkdirAll(app.dataBasePath, os.FileMode(0770)); err != nil {
		log.Panicf("Failed to create directories for app data %s: %v", app.dataBasePath, err)
		return nil
	}

	return app

}

func (a *App) initManagers() {

	// Initialize settings manager
	settingsFilePath := filepath.Join(a.configBasePath, "settings.json")
	log.Printf("Settings file path: %s", settingsFilePath)
	err := settingstore.InitSettingStore(a.settingStoreAPI, settingsFilePath)
	if err != nil {
		log.Panicf("Couldnt initialize setting store at: %s. error: %v", settingsFilePath, err)
	}

	// Initialize conversation manager
	conversationDir := filepath.Join(a.dataBasePath, "conversations")
	log.Printf("Conversation directory: %s", conversationDir)

	err = conversationstore.InitConversationCollection(a.conversationStoreAPI, conversationDir)
	if err != nil {
		log.Panicf("Couldnt initialize conversation store at: %s. err:%v", conversationDir, err)
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
	// Load the frontend
	runtime.WindowShow(a.ctx)
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

// Greet returns a greeting for the given name
func (a *App) Ping() string {
	return "pong"
}

type WrappedProviderSetAPI struct {
	*aiprovider.ProviderSetAPI
	appContext context.Context
}

// NewWrappedProviderSetAPI creates a new ProviderSet with the specified default provider
func NewWrappedProviderSetAPI(
	defaultProvider aiproviderSpec.ProviderName,
) *WrappedProviderSetAPI {
	return &WrappedProviderSetAPI{
		ProviderSetAPI: aiprovider.NewProviderSetAPI(defaultProvider),
	}
}

func (w *WrappedProviderSetAPI) setAppContext(ctx context.Context) {
	w.appContext = ctx
}

// FetchCompletion handles the completion request and streams data back to the frontend
func (w *WrappedProviderSetAPI) FetchCompletion(
	provider string,
	input aiproviderSpec.CompletionRequest,
	callbackId string,
) (*aiproviderSpec.CompletionResponse, error) {
	onStreamData := func(data string) error {
		runtime.EventsEmit(w.appContext, callbackId, data)
		return nil
	}

	resp, err := w.ProviderSetAPI.FetchCompletion(
		aiproviderSpec.ProviderName(provider),
		input,
		onStreamData,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getActualURL(origurl string) string {
	callurl := origurl
	if !strings.HasPrefix(callurl, FRONTEND_PATH_PREFIX) {
		return callurl
	}

	// percentIndex := strings.Index(callurl, "%")
	// if percentIndex != -1 {
	// 	callurl = callurl[:percentIndex]
	// }

	// qIndex := strings.Index(callurl, "?")
	// if qIndex != -1 {
	// 	callurl = callurl[:qIndex]
	// }

	// Handle if it's a page request
	if strings.HasSuffix(callurl, "/") {
		return callurl[len(FRONTEND_PATH_PREFIX):] + "index.html"
	}

	for _, d := range DIR_PAGES {
		durl := FRONTEND_PATH_PREFIX + d
		if callurl == durl {
			return callurl[len(FRONTEND_PATH_PREFIX):] + "/index.html"
		}
	}

	return callurl[len(FRONTEND_PATH_PREFIX):]
}

func LogStackTrace() {
	// Create a buffer to hold the stack trace
	buf := make([]byte, 1024)
	// Capture the stack trace
	n := goruntime.Stack(buf, false)
	// Log the stack trace
	fmt.Println("Stack trace:\n", string(buf[:n]))
}

func URLCleanerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Clean the URL using getActualURL
		// fmt.Println("Input URL:", req.URL.Path) // For debugging purposes
		// if req.URL.Path == "/agents" {
		// 	LogStackTrace()
		// }
		cleanedURL := getActualURL(req.URL.Path)
		// fmt.Println("Cleaned URL:", cleanedURL) // For debugging purposes

		// Update the request URL path
		req.URL.Path = cleanedURL

		// Call the next handler
		next.ServeHTTP(w, req)
	})
}

func EmbeddedFSWalker() {
	_ = fs.WalkDir(assets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println("Embedded file:", path)
		return nil
	})
}

func main() {
	// Create an instance of the app structure
	app := NewApp()
	app.initManagers()
	// EmbeddedFSWalker()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             AppTitle,
		Width:             1024,
		Height:            768,
		MinWidth:          1024,
		MinHeight:         768,
		MaxWidth:          7680,
		MaxHeight:         4320,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: URLCleanerMiddleware,
		},
		Menu:               nil,
		Logger:             nil,
		LogLevel:           logger.DEBUG,
		LogLevelProduction: logger.DEBUG,

		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			app.providerSetAPI.setAppContext(ctx)
		},

		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		WindowStartState: options.Normal,
		Bind: []interface{}{
			app,
			app.settingStoreAPI,
			app.conversationStoreAPI,
			app.providerSetAPI,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
			ZoomFactor:          1.0,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   AppTitle,
				Message: "",
				Icon:    icon,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
