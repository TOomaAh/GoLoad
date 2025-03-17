package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetDefaultDownloadsDirectory retourne le dossier de téléchargement par défaut
// pour l'utilisateur courant selon le système d'exploitation
func GetDefaultDownloadsDirectory() string {
	// Vérifier si nous sommes en environnement Docker
	if os.Getenv("DOCKER") == "true" {
		os.Setenv("DOWNLOAD_PATH", "/downloads")
		return "/downloads"
	}

	// Récupérer le répertoire personnel de l'utilisateur
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// En cas d'erreur, on retourne un répertoire par défaut
		return "./downloads"
	}

	// Choisir le dossier de téléchargement en fonction du système d'exploitation
	switch runtime.GOOS {
	case "windows":
		// Sur Windows, le dossier est généralement "Downloads" dans le répertoire utilisateur
		return filepath.Join(homeDir, "Downloads")
	case "darwin":
		// Sur macOS, c'est également "Downloads" dans le répertoire utilisateur
		return filepath.Join(homeDir, "Downloads")
	case "linux":
		// Sur Linux, vérifier d'abord la spécification XDG
		xdgDownloads := os.Getenv("XDG_DOWNLOAD_DIR")
		if xdgDownloads != "" {
			return xdgDownloads
		}

		// Sinon, vérifier si le fichier user-dirs.dirs existe
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome == "" {
			xdgConfigHome = filepath.Join(homeDir, ".config")
		}

		// Lire le fichier user-dirs.dirs pour trouver XDG_DOWNLOAD_DIR
		userDirsFile := filepath.Join(xdgConfigHome, "user-dirs.dirs")
		if downloadDir := getLinuxDownloadDirFromUserDirs(userDirsFile, homeDir); downloadDir != "" {
			return downloadDir
		}

		// Par défaut sur la plupart des distributions Linux
		return filepath.Join(homeDir, "Downloads")
	default:
		// Pour les autres systèmes, utiliser un dossier "Downloads" dans le répertoire utilisateur
		return filepath.Join(homeDir, "Downloads")
	}
}

// getLinuxDownloadDirFromUserDirs parse le fichier user-dirs.dirs pour extraire XDG_DOWNLOAD_DIR
func getLinuxDownloadDirFromUserDirs(userDirsPath string, homeDir string) string {
	// Vérifier que le fichier existe
	file, err := os.Open(userDirsPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Parcourir le fichier ligne par ligne
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Chercher la définition de XDG_DOWNLOAD_DIR
		if strings.HasPrefix(line, "XDG_DOWNLOAD_DIR=") {
			// Extraire le chemin entre guillemets
			value := strings.Split(line, "=")[1]
			value = strings.Trim(value, "\"'")

			// Remplacer les variables d'environnement ($HOME ou ${HOME})
			value = strings.Replace(value, "$HOME", homeDir, -1)
			value = strings.Replace(value, "${HOME}", homeDir, -1)

			return value
		}
	}

	return ""
}

// EnsureDirectoryExists s'assure qu'un répertoire existe, le crée si nécessaire
func EnsureDirectoryExists(path string) error {
	return os.MkdirAll(path, 0755)
}
