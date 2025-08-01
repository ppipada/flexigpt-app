//go:build !codeanalysis

package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/adrg/xdg"
)

const AppTitle = "FlexiGPT"

// App struct
type App struct {
	ctx                    context.Context
	settingStoreAPI        *SettingStoreWrapper
	conversationStoreAPI   *ConversationCollectionWrapper
	modelPresetStoreAPI    *ModelPresetStoreWrapper
	promptTemplateStoreAPI *PromptTemplateStoreWrapper
	toolStoreAPI           *ToolStoreWrapper
	providerSetAPI         *ProviderSetWrapper

	dataBasePath string

	settingsDirPath      string
	conversationsDirPath string
	modelPresetsDirPath  string
	promptsDirPath       string
	toolsDirPath         string
}

// NewApp creates a new App application struct
func NewApp() *App {
	if xdg.DataHome == "" {
		slog.Error(
			"Could not resolve XDG data paths",
			"XDG Data dir", xdg.DataHome,
		)
		panic("Failed to initialize App: XDG paths not set")
	}

	app := &App{}
	app.dataBasePath = filepath.Join(xdg.DataHome, strings.ToLower(AppTitle))

	app.settingsDirPath = filepath.Join(app.dataBasePath, "settings")
	app.conversationsDirPath = filepath.Join(app.dataBasePath, "conversations")
	app.modelPresetsDirPath = filepath.Join(app.dataBasePath, "modelpresets")
	app.promptsDirPath = filepath.Join(app.dataBasePath, "prompttemplates")
	app.toolsDirPath = filepath.Join(app.dataBasePath, "tools")

	if app.settingsDirPath == "" || app.conversationsDirPath == "" ||
		app.modelPresetsDirPath == "" || app.promptsDirPath == "" || app.toolsDirPath == "" {
		slog.Error(
			"Invalid app path configuration",
			"settingsDirPath", app.settingsDirPath,
			"conversationsDirPath", app.conversationsDirPath,
			"modelPresetsDirPath", app.modelPresetsDirPath,
			"promptsDirPath", app.promptsDirPath,
			"toolsDirPath", app.toolsDirPath,
		)
		panic("Failed to initialize App: invalid path configuration")
	}

	// Wails needs some instance of a struct to create bindings from its methods.
	// Therefore the pattern followed is to create a hollow struct in new and then init in startup
	app.settingStoreAPI = &SettingStoreWrapper{}
	app.conversationStoreAPI = &ConversationCollectionWrapper{}
	app.providerSetAPI = &ProviderSetWrapper{}
	app.modelPresetStoreAPI = &ModelPresetStoreWrapper{}
	app.promptTemplateStoreAPI = &PromptTemplateStoreWrapper{}
	app.toolStoreAPI = &ToolStoreWrapper{}

	if err := os.MkdirAll(app.settingsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create settings directory",
			"Settings path", app.settingsDirPath,
			"Error", err,
		)
		panic("Failed to initialize App: could not create settings directory")
	}
	if err := os.MkdirAll(app.conversationsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create conversations directory",
			"Conversations path", app.conversationsDirPath,
			"Error", err,
		)
		panic("Failed to initialize App: could not create conversations directory")
	}
	if err := os.MkdirAll(app.modelPresetsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create model presets directory",
			"Model presets path", app.modelPresetsDirPath,
			"Error", err,
		)
		panic("Failed to initialize App: could not create model presets directory")
	}
	if err := os.MkdirAll(app.promptsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create prompt templates directory",
			"Prompt Templates path", app.promptsDirPath,
			"Error", err,
		)
		panic("Failed to initialize App: could not create prompt templates directory")
	}
	if err := os.MkdirAll(app.toolsDirPath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create tools directory",
			"Tools path", app.toolsDirPath,
			"Error", err,
		)
		panic("Failed to initialize App: could not create tools directory")
	}
	slog.Info(
		"FlexiGPT paths initialized",
		"app data", app.dataBasePath,
		"settingsDirPath", app.settingsDirPath,
		"conversationsDirPath", app.conversationsDirPath,
		"modelPresetsDirPath", app.modelPresetsDirPath,
		"promptsDirPath", app.promptsDirPath,
		"toolsDirPath", app.toolsDirPath,
	)
	return app
}

func (a *App) initManagers() {
	err := InitProviderSetWrapper(a.providerSetAPI)
	if err != nil {
		slog.Error(
			"Couldn't initialize provider set",
			"Error", err,
		)
		panic("Failed to initialize managers: provider set initialization failed")
	}

	err = InitSettingStoreWrapper(a.settingStoreAPI, a.providerSetAPI, a.settingsDirPath)
	if err != nil {
		slog.Error(
			"Couldn't initialize settings store",
			"Directory", a.settingsDirPath,
			"Error", err,
		)
		panic("Failed to initialize managers: settings store initialization failed")
	}
	slog.Info("Settings store initialized", "directory", a.settingsDirPath)

	err = InitModelPresetStoreWrapper(
		a.modelPresetStoreAPI,
		a.settingStoreAPI,
		a.providerSetAPI,
		a.modelPresetsDirPath,
	)
	if err != nil {
		slog.Error(
			"Couldn't initialize model presets store",
			"Dir", a.modelPresetsDirPath,
			"Error", err,
		)
		panic("Failed to initialize managers: model presets store initialization failed")
	}
	slog.Info("Model presets store initialized", "dir", a.modelPresetsDirPath)

	err = InitConversationCollectionWrapper(a.conversationStoreAPI, a.conversationsDirPath)
	if err != nil {
		slog.Error(
			"Couldn't initialize conversation store",
			"Directory", a.conversationsDirPath,
			"Error", err,
		)
		panic("Failed to initialize managers: conversation store initialization failed")
	}
	slog.Info("Conversation store initialized", "directory", a.conversationsDirPath)

	err = InitPromptTemplateStoreWrapper(a.promptTemplateStoreAPI, a.promptsDirPath)
	if err != nil {
		slog.Error(
			"Couldn't initialize prompt template store",
			"Directory", a.promptsDirPath,
			"Error", err,
		)
		panic("Failed to initialize managers: prompt template store initialization failed")
	}

	err = InitToolStoreWrapper(a.toolStoreAPI, a.toolsDirPath)
	if err != nil {
		slog.Error(
			"Couldn't initialize tool store",
			"Directory", a.toolsDirPath,
			"Error", err,
		)
		panic("Failed to initialize managers: tool store initialization failed")
	}
}

// startup is called at application startup.
func (a *App) startup(ctx context.Context) {
	// Perform your setup here.
	a.ctx = ctx
	// Load the frontend.
	runtime.WindowShow(a.ctx)
}

// domReady is called after front-end resources have been loaded.
func (a App) domReady(ctx context.Context) {
	// Add your action here.
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination.
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here.
}

// Greet returns a greeting for the given name.
func (a *App) Ping() string {
	return "pong"
}

// SaveFile handles saving any content to a file.
func (a *App) SaveFile(
	defaultFilename string,
	contentBase64 string,
	filters []runtime.FileFilter,
) error {
	if a.ctx == nil {
		return errors.New("context is not initialized")
	}

	saveDialogOptions := runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Filters:         filters,
	}
	savePath, err := runtime.SaveFileDialog(a.ctx, saveDialogOptions)
	if err != nil {
		return err
	}
	if savePath == "" {
		// User cancelled the dialog.
		return nil
	}

	// Decode base64 content.
	contentBytes, err := base64.StdEncoding.DecodeString(contentBase64)
	if err != nil {
		return err
	}

	// Write the content to the file.
	return os.WriteFile(savePath, contentBytes, 0o644)
}
