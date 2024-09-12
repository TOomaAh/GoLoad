package main

import (
	"flag"
	"fmt"
	"gestionnaire-telechargement/internal/database"
	"gestionnaire-telechargement/internal/downloader"
	"gestionnaire-telechargement/internal/ui"
	"log"
	"os"

	"golang.org/x/text/language"
)

func main() {
	lang := flag.String("lang", "en", "Set the default language (en or fr)")
	flag.Parse()

	fmt.Println("Starting download manager")

	// Initialiser la base de données
	db, err := database.NewDatabase()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	// Définir le nombre maximum de chunks
	maxChunks := 5 // Vous pouvez ajuster cette valeur selon vos besoins

	// Initialiser le downloader
	d := downloader.NewDownloader(maxChunks)

	d.OnDownloadAdded = func(url string, totalSize int64) error {
		if err := db.AddDownload(url, totalSize); err != nil {
			return fmt.Errorf("impossible d'ajouter le téléchargement à la base de données : %v", err)
		}
		return nil
	}

	d.OnComplete = func(url string) error {
		if err := db.UpdateDownloadStatus(url, "completed"); err != nil {
			return fmt.Errorf("impossible de mettre à jour le statut du téléchargement : %v", err)
		}
		return nil
	}

	d.OnPause = func(url string) error {
		return db.UpdateDownloadStatus(url, "paused")
	}

	d.OnResume = func(url string) error {
		return db.UpdateDownloadStatus(url, "downloading")
	}

	d.OnDeleted = func(url string, deleteFile bool) error {
		db.UpdateDownloadStatus(url, "deleted")
		if deleteFile {
			details, err := db.GetDownloadDetails(url)
			if err == nil && details.SavePath != "" {
				os.Remove(details.SavePath)
			}
		}

		return db.DeleteDownload(url)
	}

	d.OnCancel = func(url string) error {
		return db.UpdateDownloadStatus(url, "cancelled")
	}

	d.OnError = func(url string, err error) {
		db.UpdateDownloadStatus(url, "failed")
	}

	// Set the default language
	switch *lang {
	case "fr":
		ui.SetLanguage(language.French)
	default:
		ui.SetLanguage(language.English)
	}

	// Initialiser l'interface utilisateur
	u := ui.NewUI(d, db)

	// Démarrer l'interface
	u.Start()
}
