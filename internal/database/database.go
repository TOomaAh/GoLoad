package database

import (
	"database/sql"
	"fmt"
	"gestionnaire-telechargement/internal/downloader"
	"log"

	_ "github.com/glebarez/go-sqlite"
)

type Database struct {
	db *sql.DB
}

type Download struct {
	ID     int64
	URL    string
	Status string
	Size   int64 // Ajoutez cette ligne
}

type Setting struct {
	Key   string
	Value string
}

func NewDatabase() (*Database, error) {

	db, err := sql.Open("sqlite", "goloader.db")
	if err != nil {
		return nil, fmt.Errorf("impossible d'ouvrir la base de données : %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à la base de données : %v", err)
	}

	database := &Database{db: db}
	if err := database.createDownloadTable(); err != nil {
		return nil, fmt.Errorf("impossible de créer la table downloads : %v", err)
	}

	// Ajoutez cette ligne pour créer la table settings
	if err := database.createSettingsTable(); err != nil {
		return nil, fmt.Errorf("impossible de créer la table settings : %v", err)
	}

	// Appelez la méthode migrate pour mettre à jour la table existante
	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("impossible de migrer la table : %v", err)
	}

	return database, nil
}

func (d *Database) createDownloadTable() error {
	query := `CREATE TABLE IF NOT EXISTS downloads (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		status TEXT NOT NULL,
		size INTEGER NOT NULL DEFAULT 0
	)`

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) createSettingsTable() error {
	query := `CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT
	)`

	_, err := d.db.Exec(query)
	return err
}

func (d *Database) AddDownload(url string, size int64) error {
	query := "INSERT INTO downloads (url, status, size) VALUES (?, ?, ?)"
	_, err := d.db.Exec(query, url, "pending", size)
	return err
}

func (d *Database) UpdateDownloadStatus(url, status string) error {
	query := "UPDATE downloads SET status = ? WHERE url = ?"
	_, err := d.db.Exec(query, status, url)
	return err
}

func (d *Database) GetPendingDownloads() ([]Download, error) {
	query := "SELECT id, url, status, size FROM downloads WHERE status = 'pending'"
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var downloads []Download
	for rows.Next() {
		var d Download
		if err := rows.Scan(&d.ID, &d.URL, &d.Status, &d.Size); err != nil {
			return nil, err
		}
		downloads = append(downloads, d)
	}

	return downloads, nil
}

func (d *Database) Close() {
	if err := d.db.Close(); err != nil {
		log.Printf("Erreur lors de la fermeture de la base de données : %v", err)
	}
}

// Ajoutez ces nouvelles méthodes

func (d *Database) GetAllDownloads() ([]Download, error) {
	query := "SELECT id, url, status, size FROM downloads"
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var downloads []Download
	for rows.Next() {
		var d Download
		if err := rows.Scan(&d.ID, &d.URL, &d.Status, &d.Size); err != nil {
			return nil, err
		}
		downloads = append(downloads, d)
	}

	return downloads, nil
}

func (d *Database) DeleteDownload(url string) error {
	query := "DELETE FROM downloads WHERE url = ?"
	_, err := d.db.Exec(query, url)
	return err
}

func (d *Database) GetDownloadByURL(url string) (Download, error) {
	query := "SELECT id, url, status, size FROM downloads WHERE url = ?"
	row := d.db.QueryRow(query, url)

	var download Download
	err := row.Scan(&download.ID, &download.URL, &download.Status, &download.Size)
	if err != nil {
		return Download{}, err
	}

	return download, nil
}

func (d *Database) migrate() error {
	// Vérifiez si la colonne 'size' existe
	query := "PRAGMA table_info(downloads)"
	rows, err := d.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var columnName string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString // Utilisez sql.NullString pour gérer les valeurs NULL
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == "size" {
			columnName = name
			break
		}
	}

	// Si la colonne 'size' n'existe pas, ajoutez-la
	if columnName != "size" {
		_, err := d.db.Exec("ALTER TABLE downloads ADD COLUMN size INTEGER NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) GetDownloadDetails(url string) (*downloader.Download, error) {
	// Implémentez la logique pour récupérer les détails du téléchargement
	// à partir de la base de données
	query := "SELECT id, url, status, size FROM downloads WHERE url = ?"
	row := db.db.QueryRow(query, url)

	var download downloader.Download
	err := row.Scan(&download.ID, &download.URL, &download.Status, &download.Size)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("aucun téléchargement trouvé pour l'URL : %s", url)
		}
		return nil, fmt.Errorf("erreur lors de la récupération des détails du téléchargement : %v", err)
	}

	return &download, nil
}

func (d *Database) GetSetting(key string) (string, error) {
	var value string
	err := d.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (d *Database) SetSetting(key, value string) error {
	_, err := d.db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}
