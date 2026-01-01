package main

import (
	"log/slog"
	"os"

	"github.com/ppipada/flexigpt-app/pkg/inferencewrapper"
	"github.com/ppipada/inference-go/debugclient"

	conversationStore "github.com/ppipada/flexigpt-app/pkg/conversation/store"
	modelpresetStore "github.com/ppipada/flexigpt-app/pkg/modelpreset/store"
	promptStore "github.com/ppipada/flexigpt-app/pkg/prompt/store"
	settingStore "github.com/ppipada/flexigpt-app/pkg/setting/store"
	toolStore "github.com/ppipada/flexigpt-app/pkg/tool/store"
)

type BackendApp struct {
	settingStoreAPI        *settingStore.SettingStore
	conversationStoreAPI   *conversationStore.ConversationCollection
	providerSetAPI         *inferencewrapper.ProviderSetAPI
	modelPresetStoreAPI    *modelpresetStore.ModelPresetStore
	promptTemplateStoreAPI *promptStore.PromptTemplateStore
	toolStoreAPI           *toolStore.ToolStore

	settingsDirPath      string
	conversationsDirPath string
	modelPresetsDirPath  string
	promptsDirPath       string
	toolsDirPath         string
}

func NewBackendApp(
	settingsDirPath, conversationsDirPath, modelPresetsDirPath, promptsDirPath, toolsDirPath string,
) *BackendApp {
	if settingsDirPath == "" || conversationsDirPath == "" ||
		modelPresetsDirPath == "" || promptsDirPath == "" || toolsDirPath == "" {
		slog.Error(
			"invalid app path configuration",
			"settingsDirPath", settingsDirPath,
			"conversationsDirPath", conversationsDirPath,
			"modelPresetsDirPath", modelPresetsDirPath,
			"promptsDirPath", promptsDirPath,
			"toolsDirPath", toolsDirPath,
		)
		panic("failed to initialize BackendApp: invalid path configuration")
	}

	app := &BackendApp{
		settingsDirPath:      settingsDirPath,
		conversationsDirPath: conversationsDirPath,
		modelPresetsDirPath:  modelPresetsDirPath,
		promptsDirPath:       promptsDirPath,
		toolsDirPath:         toolsDirPath,
	}

	app.initSettingsStore()
	app.initConversationStore()
	app.initToolStore()
	app.initProviderSet()
	app.initModelPresetStore()
	app.initPromptTemplateStore()
	return app
}

func (a *BackendApp) initSettingsStore() {
	if err := os.MkdirAll(a.settingsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"failed to create settings directory",
			"settingsDirPath", a.settingsDirPath,
			"error", err,
		)
		panic("failed to initialize BackendApp: could not create settings directory")
	}
	// Initialize settings manager.
	ss, err := settingStore.NewSettingStore(a.settingsDirPath)
	if err != nil {
		slog.Error(
			"couldn't initialize settings store",
			"settingsDirPath", a.settingsDirPath,
			"error", err,
		)
		panic("dailed to initialize BackendApp: settings store initialization failed")
	}
	a.settingStoreAPI = ss
	slog.Info("settings store initialized", "dir", a.settingsDirPath)
}

func (a *BackendApp) initConversationStore() {
	if err := os.MkdirAll(a.conversationsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"failed to create conversation store directory",
			"conversationsDirPath", a.conversationsDirPath,
			"error", err,
		)
		panic("failed to initialize BackendApp: could not create conversation store directory")
	}

	cc, err := conversationStore.NewConversationCollection(
		a.conversationsDirPath,
		conversationStore.WithFTS(true),
	)
	if err != nil {
		slog.Error(
			"couldn't initialize conversation store",
			"conversationsDirPath", a.conversationsDirPath,
			"error", err,
		)
		panic("failed to initialize BackendApp: conversation store initialization failed")
	}
	a.conversationStoreAPI = cc
	slog.Info("conversation store initialized", "directory", a.conversationsDirPath)
}

func (a *BackendApp) initModelPresetStore() {
	if err := os.MkdirAll(a.modelPresetsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"failed to create model presets directory",
			"modelPresetsDirPath", a.modelPresetsDirPath,
			"error", err,
		)
		panic("failed to initialize BackendApp: could not create model presets directory")
	}

	ms, err := modelpresetStore.NewModelPresetStore(a.modelPresetsDirPath)
	if err != nil {
		slog.Error(
			"couldn't initialize model presets store",
			"modelPresetsDirPath", a.modelPresetsDirPath,
			"error", err,
		)
		panic("failed to initialize BackendApp: model presets store initialization failed")
	}
	a.modelPresetStoreAPI = ms

	slog.Info("model presets store initialized", "filepath", a.modelPresetsDirPath)
}

func (a *BackendApp) initPromptTemplateStore() {
	if err := os.MkdirAll(a.promptsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"failed to create prompt templates directory",
			"promptsDirPath", a.promptsDirPath,
			"error", err,
		)
		panic("failed to initialize BackendApp: could not create prompt templates directory")
	}

	ps, err := promptStore.NewPromptTemplateStore(
		a.promptsDirPath,
		promptStore.WithFTS(true),
	)
	if err != nil {
		slog.Error(
			"couldn't initialize prompt template store",
			"promptsDirPath", a.promptsDirPath,
			"error", err,
		)
		panic("failed to initialize BackendApp: prompt template store initialization failed")
	}
	a.promptTemplateStoreAPI = ps
	slog.Info("prompt template store initialized", "directory", a.promptsDirPath)
}

func (a *BackendApp) initToolStore() {
	if err := os.MkdirAll(a.toolsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"failed to create tools directory",
			"toolsDirPath", a.toolsDirPath,
			"error", err,
		)
		panic("failed to initialize BackendApp: could not create tools directory")
	}

	ps, err := toolStore.NewToolStore(
		a.toolsDirPath,
		toolStore.WithFTS(true),
	)
	if err != nil {
		slog.Error(
			"couldn't initialize tools store",
			"toolsDirPath", a.toolsDirPath,
			"error", err,
		)
		panic("failed to initialize BackendApp: tool store initialization failed")
	}
	a.toolStoreAPI = ps
	slog.Info("tool store initialized", "directory", a.toolsDirPath)
}

func (a *BackendApp) initProviderSet() {
	p, err := inferencewrapper.NewProviderSetAPI(
		a.toolStoreAPI,
		inferencewrapper.WithLogger(slog.Default()),
		inferencewrapper.WithDebugConfig(&debugclient.DebugConfig{
			Disable:                 false,
			DisableRequestBody:      false,
			DisableResponseBody:     false,
			DisableContentStripping: false,
			LogToSlog:               false,
		}),
	)
	if err != nil {
		slog.Error(
			"failed to initialize provider set",
			"error", err,
		)
		panic("failed to initialize BackendApp: invalid default provider")
	}
	a.providerSetAPI = p
}
