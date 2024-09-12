package ui

import (
	"time"

	"fyne.io/fyne/v2/widget"
)

type chunkInfo struct {
	ID       int
	Size     int64
	Progress float64
}

type downloadDetails struct {
	URL            string
	Size           int64
	DownloadedSize int64
	SavePath       string
	Status         string
	Chunks         []chunkInfo
}

type downloadItem struct {
	progressBar       *widget.ProgressBar
	status            string
	speedLabel        *widget.Label
	lastUpdate        time.Time
	lastSize          float64 // Changé de int64 à float64
	pauseResumeButton *widget.Button
	url               string       // Ajoutez ce champ
	card              *widget.Card // Ajoutez ce champ
}
