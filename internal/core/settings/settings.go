package settings

import (
	utils "github.com/TOomaAh/GoLoad/internal/utils"

	"gorm.io/gorm"
)

// Settings représente les paramètres de l'application
type Settings struct {
	gorm.Model
	DarkMode     bool   `json:"darkMode"`
	DownloadPath string `json:"downloadPath"`
	MaxDownloads int    `json:"maxDownloads"`
	MaxSpeed     int64  `json:"maxSpeed"`

	AutoStartDownloads        bool `json:"autoStartDownloads"`
	RetryAfterFailure         bool `json:"retryAfterFailure"`
	NotificationEndDownload   bool `json:"notificationEndDownload"`
	NotificationErrorDownload bool `json:"notificationErrorDownload"`
}

// GetDefaultSettings retourne les paramètres par défaut
func GetDefaultSettings() Settings {
	defaultLocation := utils.GetDefaultDownloadsDirectory()

	// Créer le dossier s'il n'existe pas
	utils.EnsureDirectoryExists(defaultLocation)

	return Settings{
		DarkMode:                  true,
		DownloadPath:              defaultLocation,
		MaxDownloads:              3,
		MaxSpeed:                  0, // 0 = illimité
		AutoStartDownloads:        true,
		RetryAfterFailure:         true,
		NotificationEndDownload:   true,
		NotificationErrorDownload: true,
	}
}

// CreateOrUpdateSettings crée ou met à jour les paramètres dans la base de données
func CreateOrUpdateSettings(db *gorm.DB, settings *Settings) error {
	var existingSettings Settings
	result := db.First(&existingSettings)

	if result.Error != nil {
		// Si les paramètres n'existent pas, les créer
		return db.Create(settings).Error
	}

	// Sinon, mettre à jour les paramètres existants
	return db.Model(&existingSettings).Updates(map[string]interface{}{
		"DarkMode":                  settings.DarkMode,
		"DownloadPath":              settings.DownloadPath,
		"MaxDownloads":              settings.MaxDownloads,
		"MaxSpeed":                  settings.MaxSpeed,
		"AutoStartDownloads":        settings.AutoStartDownloads,
		"RetryAfterFailure":         settings.RetryAfterFailure,
		"NotificationEndDownload":   settings.NotificationEndDownload,
		"NotificationErrorDownload": settings.NotificationErrorDownload,
	}).Error
}

// GetSettings récupère les paramètres depuis la base de données
func GetSettings(db *gorm.DB) (*Settings, error) {
	var settings Settings
	result := db.First(&settings)

	if result.Error != nil {
		// Si les paramètres n'existent pas, retourner les paramètres par défaut
		defaultSettings := GetDefaultSettings()
		db.Create(&defaultSettings)
		return &defaultSettings, nil
	}

	return &settings, nil
}
