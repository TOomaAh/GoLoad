package ui

import (
	"fmt"
	"gestionnaire-telechargement/internal/database"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type DownloadList struct {
	ui             *UI
	container      *fyne.Container
	downloads      map[string]*downloadItem
	downloadSpeeds map[string]float64
	downloadsMutex sync.Mutex
	allDownloads   []*downloadItem // Ajoutez ce champ
	db             *database.Database
}

// Supprimez la définition de downloadItem ici

func NewDownloadList(ui *UI, db *database.Database) *DownloadList {
	dl := &DownloadList{
		ui:             ui,
		container:      container.NewVBox(),
		downloads:      make(map[string]*downloadItem),
		downloadSpeeds: make(map[string]float64),
		allDownloads:   make([]*downloadItem, 0),
		db:             db,
	}
	// Déplacez loadExistingDownloads dans une méthode séparée
	return dl
}

// Ajoutez cette nouvelle méthode
func (dl *DownloadList) Initialize() {
	dl.loadExistingDownloads()
}

func (dl *DownloadList) loadExistingDownloads() {
	downloads, err := dl.db.GetAllDownloads()
	if err != nil {
		showError(dl.ui, "Erreur", "Impossible de charger les téléchargements existants")
		return
	}

	for _, download := range downloads {
		dl.addDownloadProgressToList(download.URL, download.Status)
	}
}

func (dl *DownloadList) addDownloadProgressToList(url, status string) {
	dl.downloadsMutex.Lock()
	defer dl.downloadsMutex.Unlock()

	progress := widget.NewProgressBar()
	if status == "completed" {
		progress.SetValue(1)
	}

	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dl.deleteDownload(url)
	})
	deleteButton.Importance = widget.LowImportance

	label := widget.NewLabel(filepath.Base(url))

	detailsButton := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dl.ui.detailsPanel.showDownloadDetails(url)
	})
	detailsButton.Importance = widget.LowImportance

	pauseResumeButton := widget.NewButtonWithIcon("", theme.MediaPauseIcon(), func() {
		dl.togglePauseResume(url)
	})
	pauseResumeButton.Importance = widget.LowImportance

	speedLabel := widget.NewLabel("0 B/s")

	item := container.NewBorder(
		nil, nil,
		container.NewHBox(
			widget.NewIcon(theme.DownloadIcon()),
			label,
			speedLabel,
		),
		container.NewHBox(
			pauseResumeButton,
			detailsButton,
			deleteButton,
		),
		progress,
	)

	card := widget.NewCard("", "", item)
	dl.container.Add(card)

	downloadItem := &downloadItem{
		progressBar:       progress,
		status:            status,
		speedLabel:        speedLabel,
		lastUpdate:        time.Now(),
		lastSize:          0,
		pauseResumeButton: pauseResumeButton,
		url:               url,
		card:              card,
	}

	dl.downloads[url] = downloadItem
	dl.allDownloads = append(dl.allDownloads, downloadItem)

	dl.updatePauseResumeButton(url)
}

func (dl *DownloadList) togglePauseResume(url string) {
	dl.downloadsMutex.Lock()
	defer dl.downloadsMutex.Unlock()

	if item, exists := dl.downloads[url]; exists {
		var err error
		var action string
		if item.status == "paused" {
			err = dl.ui.downloader.ResumeDownload(url)
			action = "reprendre"
			if err == nil {
				item.status = "downloading"
			}
		} else if item.status == "downloading" || item.status == "pending" {
			err = dl.ui.downloader.PauseDownload(url)
			action = "mettre en pause"
			if err == nil {
				item.status = "paused"
			}
		}

		if err != nil {
			showError(dl.ui, "Erreur", fmt.Sprintf("Impossible de %s le téléchargement : %v", action, err))
		} else {
			dl.updatePauseResumeButton(url)
		}
	}
}

func (dl *DownloadList) updatePauseResumeButton(url string) {
	if item, exists := dl.downloads[url]; exists {
		switch item.status {
		case "paused":
			item.pauseResumeButton.SetIcon(theme.MediaPlayIcon())
			item.pauseResumeButton.Show()
		case "downloading", "pending":
			item.pauseResumeButton.SetIcon(theme.MediaPauseIcon())
			item.pauseResumeButton.Show()
		case "completed":
			item.pauseResumeButton.Hide()
		}
	}
}

func (dl *DownloadList) updateProgress(url string, progress float64) {
	dl.downloadsMutex.Lock()
	defer dl.downloadsMutex.Unlock()

	now := time.Now()

	if item, exists := dl.downloads[url]; exists {
		if item.status != "paused" {
			item.progressBar.SetValue(progress)
			if progress >= 1 {
				item.status = "completed"
				dl.updatePauseResumeButton(url)
			}

			details, err := dl.db.GetDownloadDetails(url)
			if err == nil {
				if now.Sub(item.lastUpdate) >= speedUpdateInterval {
					elapsed := now.Sub(item.lastUpdate).Seconds()
					sizeDiff := float64(details.Size) - item.lastSize

					if elapsed > 0 && sizeDiff > 0 {
						speed := sizeDiff / elapsed
						dl.downloadSpeeds[url] = speed
						item.speedLabel.SetText(formatSpeed(speed))
					}

					item.lastUpdate = now
					item.lastSize = float64(details.Size)
				}
			}
		}
	}

	dl.ui.updateGlobalSpeed()
}

func (dl *DownloadList) updateDownloadStatus(url, status string) {
	dl.downloadsMutex.Lock()
	defer dl.downloadsMutex.Unlock()

	if item, exists := dl.downloads[url]; exists {
		item.status = status
		if status == "completed" {
			item.progressBar.SetValue(1)
		}
	}
}

// Ajoutez cette nouvelle méthode
func (dl *DownloadList) filterDownloads(searchTerm, filter string) {
	dl.container.RemoveAll()
	searchTerm = strings.ToLower(searchTerm)

	dl.downloadsMutex.Lock()
	defer dl.downloadsMutex.Unlock()

	fmt.Printf("searchTerm: %s, filter: %s, allDownloads: %v\n", searchTerm, filter, dl.allDownloads)

	for _, item := range dl.allDownloads {
		if item != nil && item.card != nil {

			showItem := strings.Contains(strings.ToLower(item.url), searchTerm)

			switch filter {
			case T("inProgress"):
				showItem = showItem && (item.status == "downloading" || item.status == "pending")
			case T("completed"):
				showItem = showItem && item.status == "completed"
			case T("deleted"):
				showItem = showItem && item.status == "deleted"
			case T("errors"):
				showItem = showItem && item.status == "failed"
			}

			if showItem {
				dl.container.Add(item.card)
			}
		}
	}

	dl.container.Refresh()
}

func (dl *DownloadList) deleteDownload(url string) {
	dialog.ShowConfirm("Supprimer le téléchargement", "Voulez-vous supprimer ce téléchargement ?", func(shouldDelete bool) {
		if shouldDelete {
			err := dl.ui.downloader.SetDownloadStatusDeleted(url)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Impossible de supprimer le téléchargement : %v", err), dl.ui.window)
				return
			}
			if item, exists := dl.downloads[url]; exists {
				item.status = "deleted"
				dl.updatePauseResumeButton(url)
			}
			dl.filterDownloads("", "Tous les téléchargements")
		}
	}, dl.ui.window)
}

// Gardez cette définition et supprimez l'autre
func showDeleteConfirmDialog(u *UI, url string) {
	dialog.ShowConfirm("Supprimer le téléchargement", "Voulez-vous aussi supprimer le fichier local ?", func(deleteFile bool) {
		err := u.downloader.DeleteDownload(url, deleteFile)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Impossible de supprimer le téléchargement : %v", err), u.window)
			return
		}
		u.downloadList.removeDownloadFromList(url)
	}, u.window)
}

func (dl *DownloadList) removeDownloadFromList(url string) {
	dl.downloadsMutex.Lock()
	defer dl.downloadsMutex.Unlock()

	if item, exists := dl.downloads[url]; exists {
		dl.container.Remove(item.card)
		delete(dl.downloads, url)
	}
	dl.container.Refresh()
}

func (dl *DownloadList) updateDownloadSpeed(url string) {
	dl.downloadsMutex.Lock()
	defer dl.downloadsMutex.Unlock()

	if item, exists := dl.ui.downloads[url]; exists {
		details, err := dl.db.GetDownloadDetails(url)
		if err != nil {
			return
		}

		now := time.Now()
		if now.Sub(item.lastUpdate) >= speedUpdateInterval {
			speed := float64(details.Size) - float64(item.lastSize)/now.Sub(item.lastUpdate).Seconds()
			item.speedLabel.SetText(formatSpeed(speed))
			item.lastUpdate = now
			item.lastSize = float64(details.Size)

			dl.downloadSpeeds[url] = speed
			dl.ui.updateGlobalSpeed()
		}
	}
}
