package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/flexigpt/flexiui/pkggo/aiprovider"
	aiproviderSpec "github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
	"github.com/flexigpt/flexiui/pkggo/conversationstore"
	"github.com/flexigpt/flexiui/pkggo/logrotate"
	"github.com/flexigpt/flexiui/pkggo/settingstore"
)

// Options for the server cli.
type Options struct {
	Host                 string `doc:"Hostname to listen on."             default:"127.0.0.1"`
	Port                 int    `doc:"Port to listen on"                  default:"8888"`
	SettingsDirPath      string `doc:"path to directory of settings file"`
	ConversationsDirPath string `doc:"path to conversations directory"`
	LogsDirPath          string `doc:"path to logs directory"`
	DefaultProvider      string `doc:"default provider to use"`
	Debug                bool   `doc:"Enable debug logs"`
}

func initSlog(logsDirPath string, debug bool) *logrotate.Writer {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	slogOpts := &slog.HandlerOptions{
		Level: level,
	}

	var stdoutHandler slog.Handler = slog.NewTextHandler(os.Stdout, slogOpts)
	stdoutLogger := slog.New(stdoutHandler)

	// Init logger
	opts := logrotate.Options{
		Directory:            logsDirPath,
		MaximumFileSize:      10 * 1024 * 1024, // 10 MB
		MaximumLifetime:      24 * time.Hour,
		FileNameFunc:         logrotate.DefaultFilenameFunc,
		FlushAfterEveryWrite: true,
	}
	writer, err := logrotate.New(stdoutLogger, opts)
	if err != nil {
		slog.Error("Failed to create log writer", "Error", err)
		panic("Init logs failed")
	}

	var handler slog.Handler = slog.NewTextHandler(writer, slogOpts)
	slogger := slog.New(handler)
	slog.SetDefault(slogger)
	// slog.SetDefault(stdoutLogger)
	return writer
}

func main() {
	cli := humacli.New(func(hooks humacli.Hooks, opts *Options) {
		log.Printf("Options are %+v\n", opts)
		writer := initSlog(opts.LogsDirPath, opts.Debug)
		router := http.NewServeMux()
		api := humago.New(router, huma.DefaultConfig("FlexiGPTServer API", "1.0.0"))
		app := NewBackendApp(
			aiproviderSpec.ProviderName(opts.DefaultProvider),
			opts.SettingsDirPath,
			opts.ConversationsDirPath,
		)
		settingstore.InitSettingStoreHandlers(api, app.settingStoreAPI)
		conversationstore.InitConversationStoreHandlers(api, app.conversationStoreAPI)
		aiprovider.InitProviderSetHandlers(api, app.providerSetAPI)
		// Create the HTTP server.
		server := http.Server{
			Addr:    fmt.Sprintf("%s:%d", opts.Host, opts.Port),
			Handler: router,
		}
		hooks.OnStart(func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		})

		hooks.OnStop(func() {
			// Gracefully shutdown your server here
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			defer writer.Close()
			server.Shutdown(ctx)
		})
	})

	cli.Run()
}
