package download

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/TOomaAh/GoLoad/internal/utils"

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
	cancel      context.CancelFunc `json:"-" gorm:"-"`
	speedCalc   *SpeedCalculator   `json:"-" gorm:"-"`
	mutex       sync.RWMutex       `json:"-" gorm:"-"`
	controlChan chan string        `gorm:"-"`
}

// NewDownload crée une nouvelle instance de téléchargement
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
		speedCalc:   NewSpeedCalculator(),
		mutex:       sync.RWMutex{},
		controlChan: make(chan string, 1),
	}
}

// Lock verrouille le téléchargement pour l'accès concurrent
func (d *Download) Lock() {
	d.mutex.RLock()
}

// Unlock déverrouille le téléchargement
func (d *Download) Unlock() {
	d.mutex.RUnlock()
}

// Cancel annule le téléchargement en cours
func (d *Download) Cancel() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.cancel != nil {
		d.cancel()
	}
}

// GetSpeedCalc retourne le calculateur de vitesse
func (d *Download) GetSpeedCalc() *SpeedCalculator {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.speedCalc
}

func (d *Download) SetSpeedCalc(sc *SpeedCalculator) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.speedCalc = sc
}

// SetCancel définit la fonction d'annulation
func (d *Download) SetCancel(cancel context.CancelFunc) {
	d.Lock()
	defer d.Unlock()

	d.cancel = cancel
}

// HumanReadableSize retourne la taille formatée lisible par l'humain
func (d *Download) HumanReadableSize() string {
	return utils.FormatSize(d.Size)
}

// HumanReadableDownloaded retourne la taille téléchargée formatée
func (d *Download) HumanReadableDownloaded() string {
	return utils.FormatSize(d.Downloaded)
}

// HumanReadableSpeed retourne la vitesse formatée
func (d *Download) HumanReadableSpeed() string {
	return utils.FormatSize(d.Speed) + "/s"
}

// MarshalJSON personnalise la sérialisation JSON
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

func (d *Download) GetNameFileWithHeadRequest() (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", d.URL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	return resp.Header.Get("Content-Disposition"), nil

}
