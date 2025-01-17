package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/flexigpt/flexiui/pkggo/conversationstore"
	"github.com/flexigpt/flexiui/pkggo/settingstore"

	"github.com/flexigpt/flexiui/pkggo/aiprovider"
	aiproviderSpec "github.com/flexigpt/flexiui/pkggo/aiprovider/spec"
)

type BackendApp struct {
	settingStoreAPI      *settingstore.SettingStore
	conversationStoreAPI *conversationstore.ConversationCollection
	providerSetAPI       *aiprovider.ProviderSetAPI
	settingsDirPath      string
	settingsFilePath     string
	conversationsDirPath string
	defaultProvider      aiproviderSpec.ProviderName
}

func NewBackendApp(
	defaultProvider aiproviderSpec.ProviderName,
	settingsDirPath, conversationsDirPath string,
) *BackendApp {
	if settingsDirPath == "" || conversationsDirPath == "" || defaultProvider == "" {
		slog.Error(
			"Invalid App input",
			"settingsDirPath",
			settingsDirPath,
			"conversationsDirPath",
			conversationsDirPath,
			"defaultProvider",
			defaultProvider,
		)
		panic("Failed to initialize App")
	}

	app := &BackendApp{
		settingsDirPath:      settingsDirPath,
		conversationsDirPath: conversationsDirPath,
		defaultProvider:      defaultProvider,
	}
	app.initSettingsStore()
	app.initConversationStore()
	app.initProviderSet()
	return app
}

func (a *BackendApp) initSettingsStore() {
	a.settingStoreAPI = &settingstore.SettingStore{}
	if err := os.MkdirAll(a.settingsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create directories for settings",
			"settings dir",
			a.settingsDirPath,
			"Error",
			err,
		)
		panic("Failed to initialize App")
	}
	// Initialize settings manager
	a.settingsFilePath = filepath.Join(a.settingsDirPath, "settings.json")
	err := settingstore.InitSettingStore(a.settingStoreAPI, a.settingsFilePath)
	if err != nil {
		slog.Error(
			"Couldnt initialize setting store",
			"Settings file",
			a.settingsFilePath,
			"Error",
			err,
		)
		panic("Failed to initialize settings store")
	}
	slog.Info("Settings created", "filepath", a.settingsFilePath)
}

func (a *BackendApp) initConversationStore() {
	a.conversationStoreAPI = &conversationstore.ConversationCollection{}
	if err := os.MkdirAll(a.conversationsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create directories",
			"ConversationStore path",
			a.conversationsDirPath,
			"Error",
			err,
		)
		panic("Failed to initialize App")
	}

	err := conversationstore.InitConversationCollection(
		a.conversationStoreAPI,
		a.conversationsDirPath,
	)
	if err != nil {
		slog.Error(
			"Couldnt initialize conversation store",
			"Directory",
			a.conversationsDirPath,
			"Error",
			err,
		)
		panic("Failed to initialize Managers")
	}
	slog.Info("Conversation store initialized", "directory", a.conversationsDirPath)
}

func (a *BackendApp) initProviderSet() {
	p, err := aiprovider.NewProviderSetAPI(a.defaultProvider)
	if err != nil {
		panic("Invalid default provider")
	}
	a.providerSetAPI = p
}
