//go:build !codeanalysis

package main

import (
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

	assets "github.com/flexigpt/flexiui/packages"
	"github.com/flexigpt/flexiui/pkg/logrotate"
)

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
	// EmbeddedFSWalker(assets)

	wailsLogger := NewSlogLoggerAdapter(slogger)
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
			Assets:     assets.Assets,
			Middleware: URLCleanerMiddleware,
		},
		Menu:               nil,
		Logger:             wailsLogger,
		LogLevel:           wailsLogLevel,
		LogLevelProduction: wailsLogLevel,

		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			SetWrappedProviderAppContext(app.providerSetAPI, ctx)
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
				Icon:    assets.Icon,
			},
		},
	})
	if err != nil {
		panic(err)
	}
}
