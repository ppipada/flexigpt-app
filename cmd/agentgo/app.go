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

	modelConsts "github.com/ppipada/flexigpt-app/pkg/model/consts"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/adrg/xdg"
)

const AppTitle = "FlexiGPT"

// App struct
type App struct {
	ctx                  context.Context
	settingStoreAPI      *SettingStoreWrapper
	conversationStoreAPI *ConversationCollectionWrapper
	providerSetAPI       *ProviderSetWrapper
	modelPresetStoreAPI  *ModelPresetStoreWrapper
	configBasePath       string
	dataBasePath         string
	skillsBasePath       string
}

// NewApp creates a new App application struct
func NewApp() *App {
	if xdg.ConfigHome == "" || xdg.DataHome == "" {
		slog.Error(
			"Could not resolve data paths",
			"XDG Config dir",
			xdg.ConfigHome,
			"XDG Data dir",
			xdg.DataHome,
		)
		panic("Failed to initialize App")

	}

	app := &App{}
	app.configBasePath = filepath.Join(xdg.ConfigHome, (strings.ToLower(AppTitle)))
	app.dataBasePath = filepath.Join(xdg.DataHome, (strings.ToLower(AppTitle)))
	app.skillsBasePath = filepath.Join(app.dataBasePath, "skills")
	// Wails needs some instance of an struct to create bindings from its methods.
	// Therefore the pattern followed is that create a hollow struct in new and then init in startup
	app.settingStoreAPI = &SettingStoreWrapper{}
	app.conversationStoreAPI = &ConversationCollectionWrapper{}
	app.providerSetAPI = &ProviderSetWrapper{}
	app.modelPresetStoreAPI = &ModelPresetStoreWrapper{}

	if err := os.MkdirAll(app.configBasePath, os.FileMode(0o770)); err != nil {
		slog.Error(
			"Failed to create directories",
			"Config path",
			app.configBasePath,
			"Error",
			err,
		)
		panic("Failed to initialize app: config mkdir failed")
	}
	// This will create data dir too.
	if err := os.MkdirAll(app.skillsBasePath, os.FileMode(0o770)); err != nil {
		slog.Error("Failed to create directories", "app data", app.dataBasePath, "Error", err)
		panic("Failed to initialize app: data mkdir failed")
	}
	slog.Info(
		"FlexiGPT Paths",
		"app data",
		app.dataBasePath,
		"skills data",
		app.skillsBasePath,
		"config data",
		app.configBasePath,
	)
	return app
}

func (a *App) initManagers() {
	// Initialize settings manager
	settingsFilePath := filepath.Join(a.configBasePath, "settings.json")
	slog.Info("Setting created", "filepath", settingsFilePath)
	err := InitSettingStoreWrapper(a.settingStoreAPI, settingsFilePath)
	if err != nil {
		slog.Error(
			"Couldnt initialize setting store",
			"Setting file",
			settingsFilePath,
			"Error",
			err,
		)
		panic("Failed to initialize managers: settings store init failed")
	}

	// Initialize modelPresets manager
	modelPresetsFilePath := filepath.Join(a.skillsBasePath, "modelpresets.json")
	slog.Info("Model presets store created", "filepath", modelPresetsFilePath)
	err = InitModelPresetStoreWrapper(a.modelPresetStoreAPI, modelPresetsFilePath)
	if err != nil {
		slog.Error(
			"Couldnt initialize model presets store",
			"file",
			modelPresetsFilePath,
			"Error",
			err,
		)
		panic("Failed to initialize managers: model presets store init failed")
	}

	// Initialize conversation manager
	conversationDir := filepath.Join(a.dataBasePath, "conversations")
	slog.Info("Conversation store initialized", "directory", conversationDir)

	err = InitConversationCollectionWrapper(a.conversationStoreAPI, conversationDir)
	if err != nil {
		slog.Error(
			"Couldnt initialize conversation store",
			"Direcotry",
			conversationDir,
			"Error",
			err,
		)
		panic("Failed to initialize managers: conversations store init failed")
	}

	err = InitProviderSetWrapper(a.providerSetAPI, modelConsts.ProviderNameOpenAI)
	if err != nil {
		slog.Error(
			"Couldnt initialize providerset",
			"Error",
			err,
		)
		panic("Failed to initialize managers: provider set init failed")
	}

	err = InitProviderSetUsingSettings(a.settingStoreAPI, a.providerSetAPI)
	if err != nil {
		slog.Error(
			"Couldnt initialize providerset from settings",
			"Error",
			err,
		)
		panic("Failed to initialize managers: provider set init using settings failed")
	}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
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

// SaveFile handles saving any content to a file
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
		// User cancelled the dialog
		return nil
	}

	// Decode base64 content
	contentBytes, err := base64.StdEncoding.DecodeString(contentBase64)
	if err != nil {
		return err
	}

	// Write the content to the file
	return os.WriteFile(savePath, contentBytes, 0o644)
}
