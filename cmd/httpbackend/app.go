package main

import (
	"log/slog"
	"os"
	"path/filepath"

	conversationStore "github.com/ppipada/flexigpt-app/pkg/conversation/store"
	"github.com/ppipada/flexigpt-app/pkg/inference"
	"github.com/ppipada/flexigpt-app/pkg/settingstore"

	modelSpec "github.com/ppipada/flexigpt-app/pkg/model/spec"
	modelStore "github.com/ppipada/flexigpt-app/pkg/model/store"
	promptStore "github.com/ppipada/flexigpt-app/pkg/prompt/store"
)

type BackendApp struct {
	settingStoreAPI        *settingstore.SettingStore
	conversationStoreAPI   *conversationStore.ConversationCollection
	providerSetAPI         *inference.ProviderSetAPI
	modelPresetStoreAPI    *modelStore.ModelPresetStore
	promptTemplateStoreAPI *promptStore.PromptTemplateStore

	settingsDirPath      string
	settingsFilePath     string
	conversationsDirPath string
	modelPresetsDirPath  string
	modelPresetsFilePath string
	promptsDirPath       string

	defaultInbuiltProvider modelSpec.ProviderName
}

func NewBackendApp(
	defaultInbuiltProvider modelSpec.ProviderName,
	settingsDirPath, conversationsDirPath, modelPresetsDirPath, promptsDirPath string,
) *BackendApp {
	if settingsDirPath == "" || conversationsDirPath == "" || defaultInbuiltProvider == "" ||
		modelPresetsDirPath == "" || promptsDirPath == "" {
		slog.Error(
			"Invalid app path configuration",
			"settingsDirPath", settingsDirPath,
			"conversationsDirPath", conversationsDirPath,
			"defaultInbuiltProvider", defaultInbuiltProvider,
			"modelPresetsDirPath", modelPresetsDirPath,
			"promptsDirPath", promptsDirPath,
		)
		panic("Failed to initialize BackendApp: invalid path configuration")
	}

	app := &BackendApp{
		settingsDirPath:        settingsDirPath,
		settingsFilePath:       filepath.Join(settingsDirPath, "settings.json"),
		conversationsDirPath:   conversationsDirPath,
		modelPresetsDirPath:    modelPresetsDirPath,
		modelPresetsFilePath:   filepath.Join(modelPresetsDirPath, "modelpresets.json"),
		promptsDirPath:         promptsDirPath,
		defaultInbuiltProvider: defaultInbuiltProvider,
	}

	app.initSettingsStore()
	app.initConversationStore()
	app.initProviderSet()
	app.initModelPresetStore()
	app.initPromptTemplateStore()
	return app
}

func (a *BackendApp) initSettingsStore() {
	a.settingStoreAPI = &settingstore.SettingStore{}
	if err := os.MkdirAll(a.settingsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create settings directory",
			"settingsDirPath", a.settingsDirPath,
			"Error", err,
		)
		panic("Failed to initialize BackendApp: could not create settings directory")
	}
	// Initialize settings manager.
	err := settingstore.InitSettingStore(a.settingStoreAPI, a.settingsFilePath)
	if err != nil {
		slog.Error(
			"Couldn't initialize settings store",
			"settingsFilePath", a.settingsFilePath,
			"Error", err,
		)
		panic("Failed to initialize BackendApp: settings store initialization failed")
	}
	slog.Info("Settings store initialized", "filepath", a.settingsFilePath)
}

func (a *BackendApp) initConversationStore() {
	if err := os.MkdirAll(a.conversationsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create conversation store directory",
			"conversationsDirPath", a.conversationsDirPath,
			"Error", err,
		)
		panic("Failed to initialize BackendApp: could not create conversation store directory")
	}

	cc, err := conversationStore.NewConversationCollection(
		a.conversationsDirPath,
		conversationStore.WithFTS(true),
	)
	if err != nil {
		slog.Error(
			"Couldn't initialize conversation store",
			"conversationsDirPath", a.conversationsDirPath,
			"Error", err,
		)
		panic("Failed to initialize BackendApp: conversation store initialization failed")
	}
	a.conversationStoreAPI = cc
	slog.Info("Conversation store initialized", "directory", a.conversationsDirPath)
}

func (a *BackendApp) initModelPresetStore() {
	a.modelPresetStoreAPI = &modelStore.ModelPresetStore{}
	if err := os.MkdirAll(a.modelPresetsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create model presets directory",
			"modelPresetsDirPath", a.modelPresetsDirPath,
			"Error", err,
		)
		panic("Failed to initialize BackendApp: could not create model presets directory")
	}

	err := modelStore.InitModelPresetStore(a.modelPresetStoreAPI, a.modelPresetsFilePath)
	if err != nil {
		slog.Error(
			"Couldn't initialize model presets store",
			"modelPresetsFilePath", a.modelPresetsFilePath,
			"Error", err,
		)
		panic("Failed to initialize BackendApp: model presets store initialization failed")
	}
	slog.Info("Model presets store initialized", "filepath", a.modelPresetsFilePath)
}

func (a *BackendApp) initPromptTemplateStore() {
	if err := os.MkdirAll(a.promptsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create prompt templates directory",
			"promptsDirPath", a.promptsDirPath,
			"Error", err,
		)
		panic("Failed to initialize BackendApp: could not create prompt templates directory")
	}

	ps, err := promptStore.NewPromptTemplateStore(
		a.promptsDirPath,
		promptStore.WithFTS(true),
	)
	if err != nil {
		slog.Error(
			"Couldn't initialize prompt template store",
			"promptsDirPath", a.promptsDirPath,
			"Error", err,
		)
		panic("Failed to initialize BackendApp: prompt template store initialization failed")
	}
	a.promptTemplateStoreAPI = ps
	slog.Info("Prompt template store initialized", "directory", a.promptsDirPath)
}

func (a *BackendApp) initProviderSet() {
	p, err := inference.NewProviderSetAPI(a.defaultInbuiltProvider, false)
	if err != nil {
		slog.Error(
			"Failed to initialize provider set",
			"defaultInbuiltProvider", a.defaultInbuiltProvider,
			"Error", err,
		)
		panic("Failed to initialize BackendApp: invalid default provider")
	}
	a.providerSetAPI = p
}
