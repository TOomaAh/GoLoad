package cmd

import (
	"embed"
	"log"

	"github.com/TOomaAh/GoLoad/internal/app"
	"github.com/TOomaAh/GoLoad/internal/core/download"
	"github.com/TOomaAh/GoLoad/internal/core/settings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Run(assets *embed.FS) {
	// Initialiser la base de données
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de la base de données: %v", err)
	}

	// Créer l'application
	appInstance := app.NewApp(db)

	// Créer l'application Wails
	err = wails.Run(&options.App{
		Title:  "Gestionnaire de Téléchargements",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 240, G: 240, B: 250, A: 1},
		OnStartup:        appInstance.Startup,
		Bind: []interface{}{
			appInstance,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

// setupDatabase initialise la base de données et effectue les migrations nécessaires
func setupDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("downloads.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Appliquer les migrations pour les modèles
	err = db.AutoMigrate(&download.Download{}, &settings.Settings{})
	if err != nil {
		return nil, err
	}

	// Initialiser les paramètres par défaut s'ils n'existent pas déjà
	var count int64
	db.Model(&settings.Settings{}).Count(&count)
	if count == 0 {
		defaultSettings := settings.GetDefaultSettings()
		db.Create(&defaultSettings)
	}

	return db, nil
}
