package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ppipada/flexigpt-app/pkg/conversationstore"
	"github.com/ppipada/flexigpt-app/pkg/settingstore"

	"github.com/ppipada/flexigpt-app/pkg/aiprovider"
	aiproviderSpec "github.com/ppipada/flexigpt-app/pkg/aiprovider/spec"
)

type BackendApp struct {
	settingStoreAPI        *settingstore.SettingStore
	conversationStoreAPI   *conversationstore.ConversationCollection
	providerSetAPI         *aiprovider.ProviderSetAPI
	settingsDirPath        string
	settingsFilePath       string
	conversationsDirPath   string
	defaultInbuiltProvider aiproviderSpec.ProviderName
}

func NewBackendApp(
	defaultInbuiltProvider aiproviderSpec.ProviderName,
	settingsDirPath, conversationsDirPath string,
) *BackendApp {
	if settingsDirPath == "" || conversationsDirPath == "" || defaultInbuiltProvider == "" {
		slog.Error(
			"Invalid App input",
			"settingsDirPath",
			settingsDirPath,
			"conversationsDirPath",
			conversationsDirPath,
			"defaultInbuiltProvider",
			defaultInbuiltProvider,
		)
		panic("Failed to initialize App")
	}

	app := &BackendApp{
		settingsDirPath:        settingsDirPath,
		conversationsDirPath:   conversationsDirPath,
		defaultInbuiltProvider: defaultInbuiltProvider,
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

	cc, err := conversationstore.NewConversationCollection(
		a.conversationsDirPath,
		conversationstore.WithFTS(true),
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
	a.conversationStoreAPI = cc
	slog.Info("Conversation store initialized", "directory", a.conversationsDirPath)
}

func (a *BackendApp) initProviderSet() {
	p, err := aiprovider.NewProviderSetAPI(a.defaultInbuiltProvider, false)
	if err != nil {
		panic("Invalid default provider")
	}
	a.providerSetAPI = p
}
