package download

// StatusUpdate représente une mise à jour de statut pour les observateurs
type StatusUpdate struct {
	DownloadID uint    `json:"download_id"`
	Action     string  `json:"action"` // progress, complete, error, etc.
	Progress   float64 `json:"progress,omitempty"`
	Error      string  `json:"error,omitempty"`
}
