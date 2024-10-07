//nolint:unused
package main

import (
	"context"
	"log"
	"path/filepath"

	"github.com/flexigpt/flexiui/pkg/conversationstore"
	"github.com/flexigpt/flexiui/pkg/settingstore"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/adrg/xdg"
)

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

	configFolderPath, err := xdg.ConfigFile(AppName)
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
		log.Panicf("Couldnt initialize setting store at: %s", settingsFilePath)
	}
	a.settingsManager = sm

	// Initialize conversation manager
	conversationDir := filepath.Join(a.configBasePath, "conversations")
	log.Printf("Conversation directory: %s", conversationDir)
	cs, err := conversationstore.NewConversationCollection(conversationDir)
	if err != nil {
		log.Panicf("Couldnt initialize conversation store at: %s", conversationDir)
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
