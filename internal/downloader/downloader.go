package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ProgressCallback func(url string, progress float64)

type Downloader struct {
	DownloadDir      string
	MaxConcurrent    int // Rendu exporté
	MaxChunks        int // Ajoutez cette ligne
	semaphore        chan struct{}
	progressCallback ProgressCallback
	pausedDownloads  sync.Map
	cancelDownloads  sync.Map
	activeDownloads  sync.Map   // Ajoutez cette ligne
	mu               sync.Mutex // Ajoutez cette ligne si elle n'existe pas déjà
	OnDownloadAdded  func(url string, totalSize int64) error
	OnPause          func(url string) error
	OnResume         func(url string) error
	OnComplete       func(url string) error
	OnDeleted        func(url string, deleteFile bool) error
	OnCancel         func(url string) error
	OnUpdate         func(url string, progress float64)
	OnError          func(url string, err error)
}

type Download struct {
	ID             int64
	URL            string
	Status         string
	Size           int64
	DownloadedSize int64
	SavePath       string
	Chunks         []ChunkInfo
}

type ChunkInfo struct {
	ID       int
	Size     int64
	Progress float64
}

func NewDownloader(maxChunks int) *Downloader {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	downloadDir := filepath.Join(homeDir, "Downloads")
	maxConcurrent := 5 // Nombre maximum de téléchargements simultanés

	return &Downloader{
		DownloadDir:      downloadDir,
		MaxConcurrent:    maxConcurrent,
		MaxChunks:        maxChunks, // Initialisez MaxChunks
		semaphore:        make(chan struct{}, maxConcurrent),
		progressCallback: func(url string, progress float64) {},
		pausedDownloads:  sync.Map{},
		cancelDownloads:  sync.Map{},
		activeDownloads:  sync.Map{}, // Ajoutez cette ligne
	}
}

func (d *Downloader) SetProgressCallback(callback ProgressCallback) {
	d.progressCallback = callback
}

func (d *Downloader) Download(url string) error {
	d.semaphore <- struct{}{}        // Acquérir une place dans le sémaphore
	defer func() { <-d.semaphore }() // Libérer la place à la fin

	d.activeDownloads.Store(url, struct{}{})
	defer d.activeDownloads.Delete(url)

	// Envoyer une requête GET pour obtenir la taille du fichier
	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération des informations du fichier : %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mauvaise réponse du serveur : %s", resp.Status)
	}
	totalSize := resp.ContentLength

	// Ajouter le téléchargement à la base de données
	err = d.OnDownloadAdded(url, totalSize)

	// Créer le répertoire de téléchargement s'il n'existe pas
	if err := os.MkdirAll(d.DownloadDir, os.ModePerm); err != nil {
		return fmt.Errorf("impossible de créer le répertoire de téléchargement : %v", err)
	}

	// Obtenir le nom du fichier à partir de l'URL
	fileName := filepath.Base(url)

	// Créer le fichier de destination
	filePath := filepath.Join(d.DownloadDir, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("impossible de créer le fichier : %v", err)
	}
	defer out.Close()

	// Envoyer une requête GET pour télécharger le fichier
	resp, err = http.Get(url)
	if err != nil {
		return fmt.Errorf("erreur lors du téléchargement : %v", err)
	}
	defer resp.Body.Close()

	// Vérifier le code de statut de la réponse
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mauvaise réponse du serveur : %s", resp.Status)
	}

	// Créer un canal pour annuler le téléchargement
	cancelChan := make(chan struct{})
	d.cancelDownloads.Store(url, cancelChan)

	// Initialiser les chunks
	chunkSize := totalSize / int64(d.MaxChunks)
	chunks := make([]ChunkInfo, d.MaxChunks)
	for i := 0; i < d.MaxChunks; i++ {
		chunks[i] = ChunkInfo{
			ID:       i + 1,
			Size:     chunkSize,
			Progress: 0,
		}
	}
	// Ajuster la taille du dernier chunk
	chunks[d.MaxChunks-1].Size = totalSize - chunkSize*int64(d.MaxChunks-1)

	// Créer un lecteur qui rapporte la progression
	reader := &ProgressReader{
		Reader: resp.Body,
		Total:  totalSize,
		OnProgress: func(progress float64) {
			d.progressCallback(url, progress)
		},
	}

	// Copier le contenu du fichier
	downloaded := int64(0)
	for {
		select {
		case <-cancelChan:
			return fmt.Errorf("téléchargement annulé")
		default:
			if _, isPaused := d.pausedDownloads.Load(url); isPaused {
				time.Sleep(time.Second)
				continue
			}

			n, err := io.CopyN(out, reader, 32*1024) // Copier par blocs de 32KB
			downloaded += n
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("erreur lors de l'écriture du fichier : %v", err)
			}

			// Mettre à jour la progression
			progress := float64(downloaded) / float64(totalSize)
			d.progressCallback(url, progress)

			// Mettre à jour la progression des chunks
			for i := range chunks {
				if downloaded >= chunks[i].Size {
					chunks[i].Progress = 1
				} else {
					chunks[i].Progress = float64(downloaded) / float64(chunks[i].Size)
					break
				}
			}
		}

		if downloaded >= totalSize {
			break
		}
	}

	// Mettre à jour le statut du téléchargement dans la base de données
	d.OnComplete(url)

	d.progressCallback(url, 1.0) // Indiquer que le téléchargement est terminé
	return nil
}

type ProgressReader struct {
	io.Reader
	Total      int64
	OnProgress func(float64)
	read       int64
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.read += int64(n)
	if pr.OnProgress != nil {
		progress := float64(pr.read) / float64(pr.Total)
		pr.OnProgress(progress)
	}
	return n, err
}

func (d *Downloader) DownloadMultiple(urls []string) []error {
	var wg sync.WaitGroup
	errors := make([]error, len(urls))

	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			errors[i] = d.Download(url)
		}(i, url)
	}

	wg.Wait()
	return errors
}

func (d *Downloader) ResumePendingDownloads(pendingDownload []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	urls := make([]string, len(pendingDownload))
	for i, download := range pendingDownload {
		urls[i] = download
	}

	d.DownloadMultiple(urls)
	return nil
}

// Ajoutez cette méthode
func (d *Downloader) UpdateSemaphore() {
	d.semaphore = make(chan struct{}, d.MaxConcurrent)
}

// Ajoutez cette nouvelle méthode

func (d *Downloader) DeleteDownload(url string, deleteFile bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Arrêter le téléchargement s'il est en cours
	if cancel, exists := d.cancelDownloads.Load(url); exists {
		cancel.(chan struct{}) <- struct{}{}
		d.cancelDownloads.Delete(url)
	}

	// Supprimer de la base de données
	err := d.OnDeleted(url, deleteFile)
	if err != nil {
		return err
	}

	// Supprimer des téléchargements en cours
	d.activeDownloads.Delete(url)

	return nil
}

func (d *Downloader) PauseDownload(url string) error {
	d.pausedDownloads.Store(url, struct{}{})
	return d.OnPause(url)
}

func (d *Downloader) ResumeDownload(url string) error {
	d.pausedDownloads.Delete(url)
	err := d.OnResume(url)
	if err != nil {
		return err
	}

	// Relancer le téléchargement
	go func() {
		err := d.Download(url)
		if err != nil {
			// Gérer l'erreur (par exemple, mettre à jour le statut dans la base de données)
			d.OnError(url, err)
		}
	}()

	return nil
}

func (d *Downloader) CancelDownload(url string) error {
	if cancel, ok := d.cancelDownloads.Load(url); ok {
		close(cancel.(chan struct{}))
	}
	d.pausedDownloads.Delete(url)
	return d.OnCancel(url)
}

func (d *Downloader) SetDownloadStatusDeleted(url string) error {
	return d.OnDeleted(url, false)
}
