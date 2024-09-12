package ui

import (
	"golang.org/x/text/language"
)

var currentLang = language.English

var translations = map[language.Tag]map[string]string{
	language.English: {
		"windowTitle":               "Download Manager",
		"searchPlaceholder":         "Search downloads...",
		"globalSpeed":               "Global speed: %s",
		"filters":                   "Filters",
		"all":                       "All",
		"inProgress":                "In Progress",
		"completed":                 "Completed",
		"deleted":                   "Deleted",
		"errors":                    "Errors",
		"deleteConfirmTitle":        "Delete Download",
		"deleteConfirmMessage":      "Do you want to delete this download?",
		"deleteFileConfirmMessage":  "Do you also want to delete the local file?",
		"errorTitle":                "Error",
		"invalidURL":                "The following URL is invalid: ",
		"noValidURL":                "Please enter at least one valid URL.",
		"downloadErrorTitle":        "Download Error",
		"downloadsCompleted":        "Downloads completed",
		"downloadsCompletedMessage": "%d out of %d files downloaded successfully.",
		"settings":                  "Settings",
		"language":                  "Language",
		"english":                   "English",
		"french":                    "French",
		"close":                     "Close",
		"destinationFolder":         "Destination folder",
		"numberOfChunks":            "Number of chunks",
		"choose":                    "Choose",
		"save":                      "Save",
		"cancel":                    "Cancel",
		"enterURLs":                 "Enter URLs (one per line)",
		"addDownloadPrompt":         "Enter the URLs of the files you want to download:",
		"addDownload":               "Add Download",
		"add":                       "Add",
		"errorSavingSettings":       "Error saving settings",
		"settingsSaved":             "Settings saved",
		"settingsSavedMessage":      "Your settings have been saved successfully.",
	},
	language.French: {
		"windowTitle":               "Gestionnaire de téléchargement",
		"searchPlaceholder":         "Rechercher des téléchargements...",
		"globalSpeed":               "Vitesse globale : %s",
		"filters":                   "Filtres",
		"all":                       "Tous",
		"inProgress":                "En cours",
		"completed":                 "Terminés",
		"deleted":                   "Supprimés",
		"errors":                    "Erreurs",
		"deleteConfirmTitle":        "Supprimer le téléchargement",
		"deleteConfirmMessage":      "Voulez-vous supprimer ce téléchargement ?",
		"deleteFileConfirmMessage":  "Voulez-vous aussi supprimer le fichier local ?",
		"errorTitle":                "Erreur",
		"invalidURL":                "L'URL suivante est invalide : ",
		"noValidURL":                "Veuillez entrer au moins une URL valide.",
		"downloadErrorTitle":        "Erreur de téléchargement",
		"downloadsCompleted":        "Téléchargements terminés",
		"downloadsCompletedMessage": "%d sur %d fichiers téléchargés avec succès.",
		"settings":                  "Paramètres",
		"language":                  "Langue",
		"english":                   "Anglais",
		"french":                    "Français",
		"close":                     "Fermer",
		"destinationFolder":         "Dossier de destination",
		"numberOfChunks":            "Nombre de chunks",
		"choose":                    "Choisir",
		"save":                      "Enregistrer",
		"cancel":                    "Annuler",
		"enterURLs":                 "Entrez les URLs (une par ligne)",
		"addDownloadPrompt":         "Entrez les URLs des fichiers que vous souhaitez télécharger :",
		"addDownload":               "Ajouter un téléchargement",
		"add":                       "Ajouter",
		"errorSavingSettings":       "Erreur lors de l'enregistrement des paramètres",
		"settingsSaved":             "Paramètres enregistrés",
		"settingsSavedMessage":      "Vos paramètres ont été enregistrés avec succès.",
	},
}

func T(key string) string {
	if t, ok := translations[currentLang][key]; ok {
		return t
	}
	return key
}

func SetLanguage(lang language.Tag) {
	currentLang = lang
}
