package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/TOomaAh/GoLoad/internal/core/download"
	"github.com/TOomaAh/GoLoad/internal/core/settings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

type AppInfo struct {
	Version string
	Name    string
}

// App structure
type App struct {
	ctx     context.Context
	manager *download.Manager
	db      *gorm.DB
}

// NewApp crée une nouvelle instance de l'application
func NewApp(db *gorm.DB) *App {
	// Create application instance
	app := &App{
		manager: download.NewManager(db),
		db:      db,
	}

	// Create a channel to catch OS signals (SIGINT, SIGTERM)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Handle shutdown signals in a goroutine
	go func() {
		sig := <-c
		runtime.LogInfof(context.Background(), "Received signal: %v, shutting down gracefully", sig)

		// Close database connections and perform cleanup
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}

		// Exit with success code
		os.Exit(0)
	}()

	return app
}

// Startup est appelé au démarrage de l'application
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.manager.SetContext(&ctx)

	// Écouter les mises à jour de statut et les envoyer au frontend
	go func() {
		statusChan := a.manager.GetStatusChannel()
		for status := range statusChan {
			runtime.LogInfof(a.ctx, "Download status update: %v", status)
			// Émettre l'événement au frontend
			runtime.EventsEmit(a.ctx, "download:status", status)
		}
	}()
}

// AddDownload ajoute un nouveau téléchargement
func (a *App) AddDownload(url, filename, location string) (*download.Download, error) {
	return a.manager.AddDownload(url, filename, location)
}

// GetAllDownloads récupère tous les téléchargements
func (a *App) GetAllDownloads() []*download.Download {
	return a.manager.GetAllDownloads()
}

// GetDownloadsByStatus récupère les téléchargements par statut
func (a *App) GetDownloadsByStatus(status string) []*download.Download {
	return a.manager.GetDownloadsByStatus(status)
}

// PauseDownload met en pause un téléchargement
func (a *App) PauseDownload(id uint) error {
	return a.manager.PauseDownload(id)
}

// ResumeDownload reprend un téléchargement
func (a *App) ResumeDownload(id uint) error {
	return a.manager.ResumeDownload(id)
}

// CancelDownload annule un téléchargement
func (a *App) CancelDownload(id uint) error {
	return a.manager.CancelDownload(id)
}

// RemoveDownload supprime un téléchargement
func (a *App) RemoveDownload(id uint, deleteFile bool) error {
	return a.manager.RemoveDownload(id, deleteFile)
}

// GetSettings récupère les paramètres
func (a *App) GetSettings() settings.Settings {
	return *a.manager.GetSettings()
}

// UpdateSettings met à jour les paramètres
func (a *App) UpdateSettings(settingsObj settings.Settings) (*settings.Settings, error) {
	return a.manager.UpdateSettings(&settingsObj)
}

func (a *App) GetAppInfo() *AppInfo {
	return &AppInfo{
		Version: "1.0.0",
		Name:    "GoLoad",
	}
}
