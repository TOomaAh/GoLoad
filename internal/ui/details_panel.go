package ui

import (
	"fmt"
	"gestionnaire-telechargement/internal/database"
	"gestionnaire-telechargement/internal/downloader"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type DetailsPanel struct {
	ui               *UI
	card             *widget.Card
	container        *fyne.Container
	selectedDownload *downloader.Download // Changé de *downloadDetails à *types.Download
	urlLabel         *widget.Label        // Ajouté
	sizeLabel        *widget.Label
	downloadedLabel  *widget.Label
	statusLabel      *widget.Label
	savePathLabel    *widget.Label       // Ajouté
	progressBar      *widget.ProgressBar // Ajouté
	chunkProgressBar *ChunkProgressBar   // Ajouté
	db               *database.Database  // Ajouté
}

func NewDetailsPanel(ui *UI, db *database.Database) *DetailsPanel {
	dp := &DetailsPanel{ui: ui}
	dp.container = container.NewVBox()
	dp.card = widget.NewCard("Détails du téléchargement", "", dp.container)
	dp.card.Hide()

	// Initialiser les labels et la barre de progression
	dp.urlLabel = widget.NewLabel("")
	dp.sizeLabel = widget.NewLabel("")
	dp.downloadedLabel = widget.NewLabel("")
	dp.statusLabel = widget.NewLabel("")
	dp.savePathLabel = widget.NewLabel("")
	dp.progressBar = widget.NewProgressBar()
	dp.chunkProgressBar = NewChunkProgressBar(nil) // Assurez-vous que cette fonction existe

	return dp
}

func (dp *DetailsPanel) showDownloadDetails(url string) {
	details, err := dp.db.GetDownloadDetails(url)
	if err != nil {
		dialog.ShowError(err, dp.ui.window)
		return
	}

	dp.selectedDownload = details

	// Mettre à jour les labels avec les détails du téléchargement
	dp.urlLabel.SetText(details.URL)
	dp.sizeLabel.SetText(formatSize(details.Size))
	dp.statusLabel.SetText(formatStatus(details.Status))
	dp.savePathLabel.SetText(details.SavePath)

	// Mettre à jour la barre de progression globale
	progress := float64(details.DownloadedSize) / float64(details.Size)
	dp.progressBar.SetValue(progress)

	// Mettre à jour la barre de progression des chunks
	dp.chunkProgressBar.UpdateChunks(details.Chunks)

	dp.updateDetailsContainer()
	dp.container.Refresh()
}

func (dp *DetailsPanel) updateDetailsContainer() {
	if dp.selectedDownload == nil {
		dp.card.Hide()
		dp.setVSplitOffset(1)
		return
	}

	dp.container.RemoveAll()

	closeButton := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		dp.selectedDownload = nil
		dp.updateDetailsContainer()
	})
	closeButton.Importance = widget.LowImportance

	header := container.NewBorder(nil, nil, nil, closeButton,
		widget.NewLabel("Détails du téléchargement"))

	dp.container.Add(header)
	dp.container.Add(widget.NewSeparator())
	dp.container.Add(widget.NewLabel(fmt.Sprintf("Nom du fichier: %s", getFileName(dp.selectedDownload.URL))))
	dp.container.Add(widget.NewLabel(fmt.Sprintf("Chemin de sauvegarde: %s", dp.selectedDownload.SavePath)))

	dp.sizeLabel = widget.NewLabel(fmt.Sprintf("Taille totale: %s", formatSize(dp.selectedDownload.Size)))
	dp.container.Add(dp.sizeLabel)

	dp.downloadedLabel = widget.NewLabel(fmt.Sprintf("Téléchargé: %s", formatSize(dp.selectedDownload.DownloadedSize)))
	dp.container.Add(dp.downloadedLabel)

	dp.statusLabel = widget.NewLabel(fmt.Sprintf("Statut: %s", formatStatus(dp.selectedDownload.Status)))
	dp.container.Add(dp.statusLabel)

	// Remplacer la section des barres de progression individuelles par une seule barre de progression découpée en chunks
	if len(dp.selectedDownload.Chunks) > 0 {
		dp.container.Add(widget.NewLabel("Progression des chunks:"))
		dp.chunkProgressBar.UpdateChunks(dp.selectedDownload.Chunks)
		dp.chunkProgressBar.Resize(fyne.NewSize(200, 20)) // Définir une taille
		dp.container.Add(dp.chunkProgressBar)
	}

	dp.card.Show()
	dp.setVSplitOffset(0.7)
}

func (dp *DetailsPanel) updateProgress() {
	if dp.selectedDownload == nil {
		return
	}

	details, err := dp.db.GetDownloadDetails(dp.selectedDownload.URL)
	if err != nil {
		return
	}

	// Mettre à jour la barre de progression globale
	progress := float64(details.DownloadedSize) / float64(details.Size)
	dp.progressBar.SetValue(progress)

	// Mettre à jour la barre de progression des chunks
	dp.chunkProgressBar.UpdateChunks(details.Chunks)

	dp.container.Refresh()
}

func (dp *DetailsPanel) setVSplitOffset(offset float64) {
	if content, ok := dp.ui.window.Content().(*fyne.Container); ok {
		for _, obj := range content.Objects {
			if split, ok := obj.(*container.Split); ok {
				split.SetOffset(offset)
				break
			}
		}
	}
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatStatus(status string) string {
	switch status {
	case "pending":
		return "En attente"
	case "downloading":
		return "En cours"
	case "paused":
		return "En pause"
	case "completed":
		return "Terminé"
	case "failed":
		return "Échoué"
	default:
		return status
	}
}

func getFileName(url string) string {
	return filepath.Base(url)
}
