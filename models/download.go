package models

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// Download représente un téléchargement avec ses métadonnées
type Download struct {
	gorm.Model
	URL         string    `json:"url"`
	Filename    string    `json:"filename"`
	Location    string    `json:"location"`
	Status      string    `json:"status"` // queued, downloading, paused, completed, error, cancelled
	Progress    float64   `json:"progress"`
	Speed       int64     `json:"speed"` // octets par seconde
	Size        int64     `json:"size"`  // taille totale en octets
	Downloaded  int64     `json:"downloaded"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	ETA         string    `json:"eta,omitempty"`
	Attempt     int       `json:"attempt"`
	MaxAttempts int       `json:"max_attempts"`
	Error       string    `json:"error_message,omitempty"`

	// Champs pour le suivi interne
	cancel    context.CancelFunc `json:"-" gorm:"-"`
	speedCalc *speedCalculator   `json:"-" gorm:"-"`
	mutex     sync.RWMutex       `json:"-" gorm:"-"`
}

// Calculateur de vitesse pour les téléchargements
type speedCalculator struct {
	lastTime     time.Time
	lastBytes    int64
	currentSpeed int64
	mutex        sync.Mutex
}

func NewDownload(url, filename, location string) *Download {
	return &Download{
		URL:         url,
		Filename:    filename,
		Location:    location,
		Status:      "queued",
		Progress:    0,
		Speed:       0,
		Size:        0,
		Downloaded:  0,
		Attempt:     1,
		MaxAttempts: 3,
		speedCalc:   newSpeedCalculator(),
		mutex:       sync.RWMutex{},
	}
}

func (d *Download) Lock() {
	d.mutex.RLock()
}

func (d *Download) Unlock() {
	d.mutex.RUnlock()
}

func (d *Download) Cancel() {
	d.Lock()
	defer d.Unlock()

	if d.cancel != nil {
		d.cancel()
	}
}

func (d *Download) GetSpeedCalc() *speedCalculator {
	d.Lock()
	defer d.Unlock()

	return d.speedCalc
}

func (d *Download) SetCancel(cancel context.CancelFunc) {
	d.Lock()
	defer d.Unlock()

	d.cancel = cancel
}

func newSpeedCalculator() *speedCalculator {
	return &speedCalculator{
		lastTime:  time.Now(),
		lastBytes: 0,
	}
}

func (sc *speedCalculator) Update(totalBytes int64) int64 {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	now := time.Now()
	duration := now.Sub(sc.lastTime).Seconds()

	// Éviter une division par zéro ou des mises à jour trop fréquentes
	if duration < 0.5 {
		return sc.currentSpeed
	}

	bytesIncrement := totalBytes - sc.lastBytes
	speed := int64(float64(bytesIncrement) / duration)

	sc.currentSpeed = speed
	sc.lastTime = now
	sc.lastBytes = totalBytes

	return speed
}

// Convertit la taille en format lisible par l'humain
func (d *Download) HumanReadableSize() string {
	return formatSize(d.Size)
}

// Convertit la taille téléchargée en format lisible par l'humain
func (d *Download) HumanReadableDownloaded() string {
	return formatSize(d.Downloaded)
}

// Convertit la vitesse en format lisible par l'humain
func (d *Download) HumanReadableSpeed() string {
	return formatSize(d.Speed) + "/s"
}

// MarshalJSON pour personnaliser la sérialisation JSON
func (d *Download) MarshalJSON() ([]byte, error) {
	type Alias Download
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return json.Marshal(&struct {
		*Alias
		DownloadedReadable string `json:"downloaded_readable"`
		TotalSizeReadable  string `json:"total_size_readable"`
		SpeedReadable      string `json:"speed_readable"`
	}{
		Alias:              (*Alias)(d),
		DownloadedReadable: d.HumanReadableDownloaded(),
		TotalSizeReadable:  d.HumanReadableSize(),
		SpeedReadable:      d.HumanReadableSpeed(),
	})
}

// Fonction utilitaire pour formater les tailles en format lisible
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
