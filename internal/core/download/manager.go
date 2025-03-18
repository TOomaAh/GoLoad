package download

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/TOomaAh/GoLoad/internal/core/settings"
	"github.com/TOomaAh/GoLoad/internal/utils"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"gorm.io/gorm"
)

// Manager gère l'ensemble des téléchargements
type Manager struct {
	ctx             *context.Context
	db              *gorm.DB
	activeDownloads int
	mutex           sync.RWMutex
	statusChan      chan StatusUpdate
}

// NewManager crée une nouvelle instance du gestionnaire de téléchargements
func NewManager(db *gorm.DB) *Manager {
	// S'assurer que les paramètres existent
	settingsObj, _ := settings.GetSettings(db)

	// Créer le dossier de téléchargement s'il n'existe pas
	utils.EnsureDirectoryExists(settingsObj.DownloadPath)

	dm := &Manager{
		db:              db,
		activeDownloads: 0,
		statusChan:      make(chan StatusUpdate, 100),
	}

	// Initialiser les téléchargements actifs au démarrage
	go dm.resumeActiveDownloads()

	return dm
}

func (m *Manager) SetContext(ctx *context.Context) {
	m.ctx = ctx
}

// resumeActiveDownloads réinitialise les téléchargements actifs au démarrage
func (m *Manager) resumeActiveDownloads() {
	var downloads []Download

	// Récupérer les téléchargements en cours ou en attente
	m.db.Where("status IN ?", []string{"downloading", "queued"}).Find(&downloads)

	for i := range downloads {
		// Pour les téléchargements "downloading", les remettre en file d'attente
		if downloads[i].Status == "downloading" {
			downloads[i].Status = "queued"
			m.db.Save(&downloads[i])
		}

		// Lancer le téléchargement
		go m.StartDownload(downloads[i].ID)
	}
}

// GetSettings récupère les paramètres actuels
func (m *Manager) GetSettings() *settings.Settings {
	settingsObj, _ := settings.GetSettings(m.db)
	return settingsObj
}

// AddDownload ajoute un nouveau téléchargement
func (m *Manager) AddDownload(url, filename, location string) (*Download, error) {
	if url == "" {
		return nil, fmt.Errorf("URL de téléchargement requise")
	}

	// Extraire le nom du fichier de l'URL si non fourni
	if filename == "" {
		filename = filepath.Base(url)
		if filename == "" || filename == "." || filename == "/" {
			filename = "download"
		}
	}

	// Récupérer les paramètres pour l'emplacement par défaut
	settingsObj := m.GetSettings()

	// Utiliser l'emplacement par défaut si non spécifié
	if location == "" {
		location = settingsObj.DownloadPath
	}

	// Créer le dossier si nécessaire
	utils.EnsureDirectoryExists(location)

	// Créer un nouvel objet téléchargement
	download := NewDownload(url, filename, location)

	filename, err := download.GetNameFileWithHeadRequest()

	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération du nom du fichier: %w", err)
	}

	if filename != "" {
		download.Filename = filename
	}

	// Sauvegarder dans la base de données
	result := m.db.Create(download)
	if result.Error != nil {
		return nil, fmt.Errorf("erreur lors de la création du téléchargement: %w", result.Error)
	}

	runtime.LogInfof(*m.ctx, "Nouveau téléchargement ajouté: %s", download.Filename)
	// Démarrer automatiquement le téléchargement si le paramètre est activé
	runtime.LogInfof(*m.ctx, "AutoStartDownloads: %v", settingsObj.AutoStartDownloads)
	if settingsObj.AutoStartDownloads {
		go m.ProcessQueue()
	}

	// Envoyer une notification d'ajout
	m.statusChan <- StatusUpdate{
		DownloadID: download.ID,
		Action:     "add",
	}

	return download, nil
}

// ProcessQueue traite la file d'attente des téléchargements
func (m *Manager) ProcessQueue() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Récupérer les paramètres
	settingsObj := m.GetSettings()

	// Ne rien faire si nous avons atteint le nombre maximal de téléchargements simultanés
	if m.activeDownloads >= settingsObj.MaxDownloads {
		return
	}

	// Rechercher le prochain téléchargement en attente
	var nextDownload Download
	if result := m.db.Where("status = ? OR status = ?", "queued", "start").Order("created_at").First(&nextDownload); result.Error == nil {
		go m.StartDownload(nextDownload.ID)
	}
}

// StartDownload démarre un téléchargement spécifique
func (m *Manager) StartDownload(id uint) {

	var download Download

	if result := m.db.First(&download, id); result.Error != nil {
		log.Printf("Téléchargement avec l'ID %d non trouvé", id)
		return
	}

	// Vérifier si le téléchargement peut être démarré
	if download.Status != "queued" && download.Status != "paused" {
		return
	}

	// Mettre à jour le statut
	download.Status = "downloading"
	if download.StartTime.IsZero() {
		download.StartTime = time.Now()
	}
	m.db.Save(&download)

	m.mutex.Lock()
	m.activeDownloads++
	m.mutex.Unlock()

	// Notification de démarrage
	m.statusChan <- StatusUpdate{
		DownloadID: download.ID,
		Action:     "start",
	}

	// Créer un contexte annulable pour le téléchargement
	ctx, cancel := context.WithCancel(context.Background())

	// Stocker la fonction d'annulation
	download.SetCancel(cancel)

	// Démarrer le téléchargement dans une goroutine
	go func() {
		err := m.downloadFile(ctx, &download)
		if err != nil {
			m.onDownloadError(&download, err)
		} else {
			m.onDownloadComplete(&download)
		}
	}()
}

// downloadFile télécharge le fichier depuis l'URL
// downloadFile télécharge le fichier depuis l'URL
func (m *Manager) downloadFile(ctx context.Context, download *Download) error {
	// Créer une requête HTTP avec le contexte
	req, err := http.NewRequestWithContext(ctx, "GET", download.URL, nil)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}

	// Vérifier si nous pouvons reprendre un téléchargement existant
	tempPath := filepath.Join(download.Location, download.Filename+".part")
	var initialSize int64

	if stat, err := os.Stat(tempPath); err == nil {
		initialSize = stat.Size()
		if initialSize > 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", initialSize))
			download.Downloaded = initialSize
			m.db.Model(download).Update("downloaded", initialSize)
		}
	}

	// Effectuer la requête HTTP
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erreur de connexion: %w", err)
	}
	defer resp.Body.Close()

	// Vérifier le code de statut
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("erreur HTTP: %d", resp.StatusCode)
	}

	// Déterminer la taille totale du fichier
	var totalSize int64

	if resp.StatusCode == http.StatusPartialContent {
		contentRange := resp.Header.Get("Content-Range")
		if contentRange != "" {
			var start, end int64
			fmt.Sscanf(contentRange, "bytes %d-%d/%d", &start, &end, &totalSize)
		}
	} else {
		totalSize = resp.ContentLength
	}

	if totalSize > 0 {
		download.Size = totalSize
		m.db.Model(download).Update("size", totalSize)
	}

	// Ouvrir le fichier temporaire
	var flag int
	if initialSize > 0 {
		flag = os.O_WRONLY | os.O_APPEND
	} else {
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	file, err := os.OpenFile(tempPath, flag, 0644)
	if err != nil {
		return fmt.Errorf("erreur lors de l'ouverture du fichier: %w", err)
	}
	defer file.Close()

	// Récupérer les paramètres pour le contrôle de la vitesse
	settingsObj := m.GetSettings()

	// Télécharger le fichier
	buffer := make([]byte, 64*1024) // Buffer de 64KB
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Canal done pour signaler la fermeture de la goroutine de contrôle
	done := make(chan struct{})
	defer close(done)

	// Goroutine pour écouter les commandes de contrôle
	go func() {
		for {
			select {
			case cmd, ok := <-download.controlChan:
				if !ok {
					// Canal fermé, sortir
					return
				}
				if cmd == "pause" {
					log.Printf("Commande de pause reçue pour le téléchargement %d", download.ID)

					// Le statut est déjà mis à jour dans PauseDownload
					// et le contexte sera annulé par le cancel(), ce qui arrêtera la routine principale

					// On retourne nil depuis la routine principale pour indiquer une pause propre
					return
				}
			case <-done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Contexte annulé pour le téléchargement %d: %v", download.ID, ctx.Err())

			// Vérifier si c'est une pause ou une annulation
			var currentStatus string
			if result := m.db.Model(&Download{}).Where("id = ?", download.ID).Select("status").Scan(&currentStatus); result.Error == nil {
				if currentStatus == "paused" {
					return nil // Retourner nil pour une pause (pas d'erreur)
				}
			}

			return ctx.Err()

		case <-ticker.C:
			// Vérifier le statut actuel dans la base de données (une fois par tick)
			var currentStatus string
			m.db.Model(&Download{}).Where("id = ?", download.ID).Select("status").Scan(&currentStatus)
			if currentStatus == "paused" {
				return nil
			}
			if currentStatus == "cancelled" {
				return fmt.Errorf("téléchargement annulé")
			}

			// Calculer la vitesse
			if download.GetSpeedCalc() == nil {
				download.SetSpeedCalc(NewSpeedCalculator())
			}
			download.Speed = download.GetSpeedCalc().Update(download.Downloaded)

			// Calculer l'ETA (temps restant estimé)
			var eta string
			if download.Speed > 0 && download.Size > download.Downloaded {
				remainingBytes := download.Size - download.Downloaded
				etaSeconds := remainingBytes / download.Speed
				etaMinutes := etaSeconds / 60
				etaRemainingSeconds := etaSeconds % 60
				eta = fmt.Sprintf("%02d:%02d", etaMinutes, etaRemainingSeconds)
			} else {
				eta = "--:--"
			}

			// Calculer la progression
			var progress float64
			if download.Size > 0 {
				progress = float64(download.Downloaded) / float64(download.Size) * 100
			}

			// Mettre à jour toutes les statistiques en une seule fois
			m.db.Model(download).Updates(map[string]interface{}{
				"speed":      download.Speed,
				"eta":        eta,
				"downloaded": download.Downloaded,
				"progress":   progress,
			})

			// Envoyer une mise à jour du statut au frontend
			m.statusChan <- StatusUpdate{
				DownloadID: download.ID,
				Action:     "progress",
				Progress:   progress,
			}

		default:
			// Lire les données
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				// Écrire dans le fichier
				_, writeErr := file.Write(buffer[:n])
				if writeErr != nil {
					return fmt.Errorf("erreur d'écriture fichier: %w", writeErr)
				}

				// Mettre à jour la progression en mémoire seulement
				// (pas d'appel BDD ici)
				download.Downloaded += int64(n)

				// Limiter la vitesse si nécessaire
				if settingsObj.MaxSpeed > 0 {
					currentSpeed := download.Speed
					if currentSpeed > settingsObj.MaxSpeed {
						delay := float64(n) / float64(settingsObj.MaxSpeed)
						time.Sleep(time.Duration(delay * float64(time.Second)))
					}
				}
			}

			// Gérer la fin du téléchargement ou les erreurs
			if err != nil {
				if err == io.EOF {
					// Téléchargement terminé
					finalPath := filepath.Join(download.Location, download.Filename)
					file.Close()
					os.Rename(tempPath, finalPath)

					// Mise à jour finale à 100%
					m.db.Model(download).Updates(map[string]interface{}{
						"downloaded": download.Size,
						"progress":   100.0,
						"status":     "completed",
						"end_time":   time.Now(),
						"speed":      0,
						"eta":        "00:00",
					})

					return nil
				}
				return fmt.Errorf("erreur de lecture: %w", err)
			}
		}
	}
}

// onDownloadComplete est appelé lorsqu'un téléchargement est terminé
// onDownloadComplete est appelé lorsqu'un téléchargement est terminé
func (m *Manager) onDownloadComplete(download *Download) {
	// Fermer le canal de contrôle pour éviter les fuites de goroutines
	download.mutex.Lock()
	if download.controlChan != nil {
		close(download.controlChan)
		download.controlChan = nil
	}
	download.mutex.Unlock()

	// Mettre à jour le téléchargement dans la base de données
	m.db.Model(download).Updates(map[string]interface{}{
		"status":   "completed",
		"end_time": time.Now(),
		"progress": 100,
		"speed":    0,
		"eta":      "00:00",
	})

	m.mutex.Lock()
	m.activeDownloads--
	m.mutex.Unlock()

	// Envoyer une notification de fin
	m.statusChan <- StatusUpdate{
		DownloadID: download.ID,
		Action:     "complete",
	}

	// Traiter la file d'attente pour démarrer un autre téléchargement
	go m.ProcessQueue()
}

// onDownloadError est appelé en cas d'erreur pendant le téléchargement
func (m *Manager) onDownloadError(download *Download, err error) {
	log.Printf("Erreur lors du téléchargement de %s: %v", download.Filename, err)

	// Fermer le canal de contrôle
	download.mutex.Lock()
	if download.controlChan != nil && err != context.Canceled {
		close(download.controlChan)
		download.controlChan = nil
	}
	download.mutex.Unlock()

	// Si l'erreur est due à une annulation, ne pas réessayer
	if err == context.Canceled {
		m.db.Model(download).Update("status", "cancelled")

		m.mutex.Lock()
		m.activeDownloads--
		m.mutex.Unlock()

		go m.ProcessQueue()
		return
	}

	// Récupérer les paramètres pour savoir si on doit réessayer
	settingsObj := m.GetSettings()

	// Incrémenter le nombre de tentatives
	download.Attempt++

	// Si nous n'avons pas dépassé le nombre maximum de tentatives et que l'option est activée
	if settingsObj.RetryAfterFailure && download.Attempt < download.MaxAttempts {
		// Mettre à jour dans la base de données
		m.db.Model(download).Updates(map[string]interface{}{
			"status":  "queued",
			"attempt": download.Attempt,
			"error":   fmt.Sprintf("Tentative %d/%d: %v", download.Attempt, download.MaxAttempts, err),
		})

		log.Printf("Nouvelle tentative (%d/%d) pour %s", download.Attempt, download.MaxAttempts, download.Filename)

		// Attendre avant de réessayer (backoff exponentiel)
		delay := time.Duration(1<<uint(download.Attempt-1)) * time.Second
		time.Sleep(delay)

		// Réessayer
		go m.StartDownload(download.ID)
	} else {
		// Marquer comme échoué après plusieurs tentatives
		m.db.Model(download).Updates(map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		})

		m.mutex.Lock()
		m.activeDownloads--
		m.mutex.Unlock()

		// Envoyer une notification d'erreur
		m.statusChan <- StatusUpdate{
			DownloadID: download.ID,
			Action:     "error",
			Error:      err.Error(),
		}

		// Traiter la file d'attente
		go m.ProcessQueue()
	}
}

// PauseDownload met en pause un téléchargement
func (m *Manager) PauseDownload(id uint) error {
	var download Download
	if result := m.db.First(&download, id); result.Error != nil {
		return fmt.Errorf("téléchargement avec l'ID %d non trouvé", id)
	}

	if download.Status != "downloading" {
		return fmt.Errorf("impossible de mettre en pause: statut actuel est %s", download.Status)
	}

	// Envoyer la commande de pause à la goroutine de téléchargement
	download.mutex.Lock() // Utilisez le mutex standard
	if download.controlChan != nil {
		select {
		case download.controlChan <- "pause":
			// La commande a été envoyée
			log.Printf("Commande de pause envoyée pour le téléchargement %d", download.ID)
		default:
			// Le canal est plein ou fermé, mise à jour directe de la BDD
			log.Printf("Canal de contrôle non disponible, mise à jour directe du statut pour le téléchargement %d", download.ID)
			m.db.Model(&download).Updates(map[string]interface{}{
				"status": "paused",
				"speed":  0,
				"eta":    "--:--",
			})

			m.mutex.Lock()
			m.activeDownloads--
			m.mutex.Unlock()
		}
	}
	download.mutex.Unlock()

	return nil
}

// ResumeDownload reprend un téléchargement en pause
func (m *Manager) ResumeDownload(id uint) error {
	var download Download
	if result := m.db.First(&download, id); result.Error != nil {
		return fmt.Errorf("téléchargement avec l'ID %d non trouvé", id)
	}

	if download.Status != "paused" {
		return fmt.Errorf("impossible de reprendre: statut actuel est %s", download.Status)
	}

	// Mettre à jour dans la base de données
	m.db.Model(&download).Update("status", "queued")

	// Notifier de la reprise
	m.statusChan <- StatusUpdate{
		DownloadID: download.ID,
		Action:     "resume",
	}

	// Traiter la file d'attente
	go m.ProcessQueue()

	return nil
}

// CancelDownload annule un téléchargement
func (m *Manager) CancelDownload(id uint) error {
	var download Download
	if result := m.db.First(&download, id); result.Error != nil {
		return fmt.Errorf("téléchargement avec l'ID %d non trouvé", id)
	}

	if download.Status != "downloading" && download.Status != "paused" && download.Status != "queued" {
		return fmt.Errorf("impossible d'annuler: statut actuel est %s", download.Status)
	}

	wasActive := download.Status == "downloading"

	// Mettre à jour dans la base de données
	m.db.Model(&download).Update("status", "cancelled")

	// Annuler l'opération en cours si elle est active
	download.Cancel()

	// Supprimer le fichier partiel
	tempPath := filepath.Join(download.Location, download.Filename+".part")
	os.Remove(tempPath)

	if wasActive {
		m.mutex.Lock()
		m.activeDownloads--
		m.mutex.Unlock()

		// Traiter la file d'attente
		go m.ProcessQueue()
	}

	// Envoyer une notification d'annulation
	m.statusChan <- StatusUpdate{
		DownloadID: download.ID,
		Action:     "cancel",
	}

	return nil
}

// RemoveDownload supprime un téléchargement de la liste
func (m *Manager) RemoveDownload(id uint, deleteFile bool) error {
	var download Download
	if result := m.db.First(&download, id); result.Error != nil {
		return fmt.Errorf("téléchargement avec l'ID %d non trouvé", id)
	}

	// Si le téléchargement est actif, l'annuler d'abord
	wasActive := download.Status == "downloading"
	if wasActive {
		download.Cancel()
	}

	// Supprimer le fichier si demandé
	if deleteFile {
		filePath := filepath.Join(download.Location, download.Filename)
		os.Remove(filePath)

		tempPath := filepath.Join(download.Location, download.Filename+".part")
		os.Remove(tempPath)
	}

	// Supprimer de la base de données
	m.db.Delete(&download)

	if wasActive {
		m.mutex.Lock()
		m.activeDownloads--
		m.mutex.Unlock()

		go m.ProcessQueue()
	}

	// Envoyer une notification de suppression
	m.statusChan <- StatusUpdate{
		DownloadID: download.ID,
		Action:     "remove",
	}

	return nil
}

// GetAllDownloads retourne tous les téléchargements
func (m *Manager) GetAllDownloads() []*Download {
	var downloads []*Download
	m.db.Find(&downloads)
	return downloads
}

// GetDownloadsByStatus filtre les téléchargements par statut
func (m *Manager) GetDownloadsByStatus(status string) []*Download {
	var downloads []*Download
	m.db.Where("status = ?", status).Find(&downloads)
	return downloads
}

// UpdateSettings met à jour les paramètres
func (m *Manager) UpdateSettings(settingsObj *settings.Settings) (*settings.Settings, error) {
	// Mettre à jour la base de données
	err := settings.CreateOrUpdateSettings(m.db, settingsObj)
	if err != nil {
		return nil, err
	}

	// Créer le dossier de téléchargement s'il n'existe pas
	utils.EnsureDirectoryExists(settingsObj.DownloadPath)

	return settingsObj, nil
}

// GetStatusChannel retourne le canal pour les mises à jour de statut
func (m *Manager) GetStatusChannel() <-chan StatusUpdate {
	return m.statusChan
}
