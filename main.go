package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"context"

	"path/filepath"
	"strings"

	"github.com/flexigpt/flexiui/pkg/conversationstore"
	"github.com/flexigpt/flexiui/pkg/settingstore"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/adrg/xdg"
)

const AppTitle = "FlexiGptUI"

//go:embed all:frontend/dist
var assets embed.FS

//go:embed images/icon.png
var icon []byte

// App struct
type App struct {
	ctx                 context.Context
	settingsManager     *settingstore.SettingsStore
	conversationManager *conversationstore.ConversationCollection
	// providerSetManager  aiproviderSpec.ProviderSetAPI
	configBasePath string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx

	configFolderPath, err := xdg.ConfigFile(strings.ToLower(AppTitle))
	if err != nil {
		log.Panicf("Could not resolve path for config dir: %v", err)
		return
	}

	a.configBasePath = configFolderPath
	// Initialize settings manager
	settingsFilePath := filepath.Join(a.configBasePath, "settings.json")
	log.Printf("Settings file path: %s", settingsFilePath)
	sm, err := settingstore.NewSettingStore(settingsFilePath)
	if err != nil {
		log.Panicf("Couldnt initialize setting store at: %s. error: %v", settingsFilePath, err)
	}
	a.settingsManager = sm

	// Initialize conversation manager
	conversationDir := filepath.Join(a.configBasePath, "conversations")
	log.Printf("Conversation directory: %s", conversationDir)
	cs, err := conversationstore.NewConversationCollection(conversationDir)
	if err != nil {
		log.Panicf("Couldnt initialize conversation store at: %s. err:%v", conversationDir, err)
	}
	a.conversationManager = cs
	// Initialize provider set manager
	// a.providerSetManager = NewProviderSet("OPENAI")

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

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             AppTitle,
		Width:             1024,
		Height:            768,
		MinWidth:          1024,
		MinHeight:         768,
		MaxWidth:          1280,
		MaxHeight:         800,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:             nil,
		Logger:           nil,
		LogLevel:         logger.DEBUG,
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		WindowStartState: options.Normal,
		Bind: []interface{}{
			app,
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
