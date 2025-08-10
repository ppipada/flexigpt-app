//go:build !codeanalysis

package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	assets "github.com/ppipada/flexigpt-app/frontend"
	"github.com/ppipada/flexigpt-app/pkg/logrotate"
)

var Version string

func main() {
	appDisplayTitle := "FlexiGPT - " + Version
	slogLevel := slog.LevelDebug
	wailsLogLevel := logger.INFO
	wailsProdLogLevel := logger.INFO

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
	slog.Info("Initialized", "App", appDisplayTitle)

	// Create application with options
	err = wails.Run(&options.App{
		Title:             appDisplayTitle,
		MinWidth:          1024,
		MinHeight:         768,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets:     assets.Assets,
			Middleware: URLCleanerMiddleware,
		},
		Menu:                     nil,
		Logger:                   wailsLogger,
		LogLevel:                 wailsLogLevel,
		LogLevelProduction:       wailsProdLogLevel,
		EnableDefaultContextMenu: true,
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			SetWrappedProviderAppContext(app.providerSetAPI, ctx)
		},

		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		WindowStartState: options.Fullscreen,
		Bind: []any{
			app,
			app.settingStoreAPI,
			app.conversationStoreAPI,
			app.providerSetAPI,
			app.modelPresetStoreAPI,
			app.promptTemplateStoreAPI,
			app.toolStoreAPI,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: mac.TitleBarDefault(),
			About: &mac.AboutInfo{
				Title:   appDisplayTitle,
				Message: "An AI app platform.\n\nCopyright © 2023 - Present",
				Icon:    assets.Icon,
			},
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},

		Linux: &linux.Options{
			ProgramName:         appDisplayTitle,
			Icon:                assets.Icon,
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyOnDemand,
		},
	})
	if err != nil {
		panic(err)
	}
}
