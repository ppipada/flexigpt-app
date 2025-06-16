package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ppipada/flexigpt-app/pkg/conversationstore"
	"github.com/ppipada/flexigpt-app/pkg/settingstore"

	"github.com/ppipada/flexigpt-app/pkg/inference"
	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"
	modelStore "github.com/ppipada/flexigpt-app/pkg/model/store"
)

type BackendApp struct {
	settingStoreAPI        *settingstore.SettingStore
	conversationStoreAPI   *conversationstore.ConversationCollection
	providerSetAPI         *inference.ProviderSetAPI
	modelPresetsStoreAPI   *modelStore.ModelPresetsStore
	settingsDirPath        string
	settingsFilePath       string
	conversationsDirPath   string
	skillsDirPath          string
	modelPresetsFilePath   string
	defaultInbuiltProvider modelSpec.ProviderName
}

func NewBackendApp(
	defaultInbuiltProvider modelSpec.ProviderName,
	settingsDirPath, conversationsDirPath, skillsDirPath string,
) *BackendApp {
	if settingsDirPath == "" || conversationsDirPath == "" || defaultInbuiltProvider == "" ||
		skillsDirPath == "" {
		slog.Error(
			"Invalid App input",
			"settingsDirPath",
			settingsDirPath,
			"conversationsDirPath",
			conversationsDirPath,
			"defaultInbuiltProvider",
			defaultInbuiltProvider,
			"skillsDirPath",
			skillsDirPath,
		)
		panic("Failed to initialize App")
	}

	app := &BackendApp{
		settingsDirPath:        settingsDirPath,
		conversationsDirPath:   conversationsDirPath,
		defaultInbuiltProvider: defaultInbuiltProvider,
		skillsDirPath:          skillsDirPath,
	}
	app.initSettingsStore()
	app.initConversationStore()
	app.initProviderSet()
	app.initModelPresetsStore()
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
	// Initialize settings manager.
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

func (a *BackendApp) initModelPresetsStore() {
	a.modelPresetsStoreAPI = &modelStore.ModelPresetsStore{}
	if err := os.MkdirAll(a.skillsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create directories for skills",
			"skills dir",
			a.skillsDirPath,
			"Error",
			err,
		)
		panic("Failed to initialize App")
	}
	// Initialize model presets store.
	a.modelPresetsFilePath = filepath.Join(a.skillsDirPath, "modelpresets.json")
	err := modelStore.InitModelPresetsStore(a.modelPresetsStoreAPI, a.modelPresetsFilePath)
	if err != nil {
		slog.Error(
			"Couldnt initialize model presets store",
			"Presets file",
			a.modelPresetsFilePath,
			"Error",
			err,
		)
		panic("Failed to initialize model presets store")
	}
	slog.Info("Model Presets store created", "filepath", a.modelPresetsFilePath)
}

func (a *BackendApp) initProviderSet() {
	p, err := inference.NewProviderSetAPI(a.defaultInbuiltProvider, false)
	if err != nil {
		panic("Invalid default provider")
	}
	a.providerSetAPI = p
}
