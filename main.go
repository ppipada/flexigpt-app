//go:build !codeanalysis

package main

import (
	"embed"
	"log/slog"
	"os"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"context"

	"path/filepath"
	"strings"

	"github.com/flexigpt/flexiui/pkg/aiprovider/openai"
	"github.com/flexigpt/flexiui/pkg/conversationstore"
	"github.com/flexigpt/flexiui/pkg/logrotate"
	"github.com/flexigpt/flexiui/pkg/settingstore"
	"github.com/flexigpt/flexiui/pkg/wailsutils"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/adrg/xdg"
)

const AppTitle = "FlexiGPT"
const AppDisplayTitle = "FlexiGPT"

//go:embed all:packages/frontend/build
var assets embed.FS

//go:embed packages/frontend/public/icon.png
var icon []byte

// App struct
type App struct {
	ctx                  context.Context
	settingStoreAPI      *settingstore.SettingStore
	conversationStoreAPI *conversationstore.ConversationCollection
	providerSetAPI       *wailsutils.WrappedProviderSetAPI
	configBasePath       string
	dataBasePath         string
}

// NewApp creates a new App application struct
func NewApp() *App {
	if xdg.ConfigHome == "" || xdg.DataHome == "" {
		slog.Error(
			"Could not resolve data paths",
			"XDG Config dir",
			xdg.ConfigHome,
			"XDG Data dir",
			xdg.DataHome,
		)
		panic("Failed to initialize App")

	}

	app := &App{}
	app.configBasePath = filepath.Join(xdg.ConfigHome, (strings.ToLower(AppTitle)))
	app.dataBasePath = filepath.Join(xdg.DataHome, (strings.ToLower(AppTitle)))
	app.settingStoreAPI = &settingstore.SettingStore{}
	app.conversationStoreAPI = &conversationstore.ConversationCollection{}
	app.providerSetAPI = wailsutils.NewWrappedProviderSetAPI(openai.ProviderNameOpenAI)

	if err := os.MkdirAll(app.configBasePath, os.FileMode(0770)); err != nil {
		slog.Error(
			"Failed to create directories",
			"Config path",
			app.configBasePath,
			"Error",
			err,
		)
		panic("Failed to initialize App")
	}
	if err := os.MkdirAll(app.dataBasePath, os.FileMode(0770)); err != nil {
		slog.Error("Failed to create directories", "app data", app.dataBasePath, "Error", err)
		panic("Failed to initialize App")
	}

	return app

}

func (a *App) initManagers() {

	// Initialize settings manager
	settingsFilePath := filepath.Join(a.configBasePath, "settings.json")
	slog.Info("Settings created", "filepath", settingsFilePath)
	err := settingstore.InitSettingStore(a.settingStoreAPI, settingsFilePath)
	if err != nil {
		slog.Error(
			"Couldnt initialize setting store",
			"Settings file",
			settingsFilePath,
			"Error",
			err,
		)
		panic("Failed to initialize Managers")
	}

	// Initialize conversation manager
	conversationDir := filepath.Join(a.dataBasePath, "conversations")
	slog.Info("Conversation store initialized", "directory", conversationDir)

	err = conversationstore.InitConversationCollection(a.conversationStoreAPI, conversationDir)
	if err != nil {
		slog.Error(
			"Couldnt initialize conversation store",
			"Direcotry",
			conversationDir,
			"Error",
			err,
		)
		panic("Failed to initialize Managers")
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

func main() {
	slogLevel := slog.LevelDebug
	wailsLogLevel := logger.DEBUG

	// Create an instance of the app structure
	app := NewApp()
	slogOpts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	var stdoutHandler slog.Handler = slog.NewTextHandler(os.Stdout, slogOpts)
	stdoutLogger := slog.New(stdoutHandler)

	// Init logger
	opts := logrotate.Options{
		Directory:            filepath.Join(app.dataBasePath, "logs"),
		MaximumFileSize:      10 * 1024 * 1024, // 10 MB
		MaximumLifetime:      24 * time.Hour,
		FileNameFunc:         logrotate.DefaultFilenameFunc,
		FlushAfterEveryWrite: true,
	}
	writer, err := logrotate.New(stdoutLogger, opts)
	if err != nil {
		slog.Error("Failed to create log writer", "Error", err)
		panic("Init failed")
	}
	defer writer.Close()

	var handler slog.Handler = slog.NewTextHandler(writer, slogOpts)
	slogger := slog.New(handler)
	slog.SetDefault(slogger)

	app.initManagers()
	// wailsutils.EmbeddedFSWalker(assets)

	wailsLogger := wailsutils.NewSlogLoggerAdapter(slogger)
	slog.Info("App Initialized")
	// Create application with options
	err = wails.Run(&options.App{
		Title:             AppDisplayTitle,
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
			Middleware: wailsutils.URLCleanerMiddleware,
		},
		Menu:               nil,
		Logger:             wailsLogger,
		LogLevel:           wailsLogLevel,
		LogLevelProduction: wailsLogLevel,

		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			wailsutils.SetWrappedProviderAppContext(app.providerSetAPI, ctx)
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
				Title:   AppDisplayTitle,
				Message: "",
				Icon:    icon,
			},
		},
	})
	if err != nil {
		panic(err)
	}
}
