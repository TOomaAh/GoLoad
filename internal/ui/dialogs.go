package ui

import (
	"fmt"
	"net/url"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showAddDownloadDialog(u *UI) {
	clipboardContent := u.getClipboardContent()

	urlEntry := widget.NewMultiLineEntry()
	urlEntry.SetPlaceHolder("Entrez les URLs à télécharger (une par ligne)")

	if clipboardContent != "" && isURL(clipboardContent) {
		urlEntry.SetText(clipboardContent)
	}

	pathEntry := widget.NewEntry()
	pathEntry.SetText(u.downloader.DownloadDir)

	pathButton := widget.NewButton("Choisir", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, u.window)
				return
			}
			if uri == nil {
				return
			}
			pathEntry.SetText(uri.Path())
		}, u.window)
	})
	pathButton.Importance = widget.HighImportance

	pathContainer := container.NewBorder(nil, nil, nil, pathButton, pathEntry)

	content := container.NewVBox(
		widget.NewLabel("URLs à télécharger :"),
		urlEntry,
		widget.NewLabel("Chemin de sauvegarde :"),
		pathContainer,
	)

	dialog.ShowCustomConfirm("Ajouter des téléchargements", "Télécharger", "Annuler", content, func(download bool) {
		if download {
			u.downloader.DownloadDir = pathEntry.Text
			go u.downloadMultiple(urlEntry.Text) // Modifié ici
		}
	}, u.window)
}

func isURL(s string) bool {
	if _, err := url.ParseRequestURI(s); err != nil {
		return false
	}
	return true
}

func showSettingsDialog(u *UI) {
	pathEntry := widget.NewEntry()
	pathEntry.SetText(u.downloader.DownloadDir)

	chunkEntry := widget.NewEntry()
	chunkEntry.SetText(strconv.Itoa(u.downloader.MaxConcurrent))

	content := container.NewVBox(
		widget.NewLabel("Chemin de sauvegarde :"),
		pathEntry,
		widget.NewLabel("Nombre de téléchargements simultanés :"),
		chunkEntry,
	)

	dialog.ShowCustomConfirm("Paramètres", "Enregistrer", "Annuler", content, func(save bool) {
		if save {
			u.downloader.DownloadDir = pathEntry.Text
			maxConcurrent, _ := strconv.Atoi(chunkEntry.Text)
			u.downloader.MaxConcurrent = maxConcurrent
			u.downloader.UpdateSemaphore()
		}
	}, u.window)
}

func showError(u *UI, title, message string) {
	dialog.ShowError(fmt.Errorf(message), u.window)
}

func showInfo(u *UI, title, message string) {
	dialog.ShowInformation(title, message, u.window)
}

func resumeDownloads(u *UI) {
	pendings, err := u.db.GetPendingDownloads()

	if err != nil {
		showError(u, "Erreur de reprise", fmt.Sprintf("Impossible de récupérer les téléchargements en attente : %v", err))
		return
	}

	pendingUrls := make([]string, len(pendings))
	for i, download := range pendings {
		pendingUrls[i] = download.URL
	}

	err = u.downloader.ResumePendingDownloads(pendingUrls)
	if err != nil {
		showError(u, "Erreur de reprise", fmt.Sprintf("Impossible de reprendre les téléchargements : %v", err))
	} else {
		showInfo(u, "Reprise des téléchargements", "Les téléchargements en attente ont été repris.")
	}
}
