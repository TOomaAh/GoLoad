package ui

import (
	"errors"
	"fmt"
	"gestionnaire-telechargement/internal/database"
	"gestionnaire-telechargement/internal/downloader"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/text/language"
)

const (
	speedUpdateInterval = 500 * time.Millisecond
)

type UI struct {
	downloader       *downloader.Downloader
	window           fyne.Window
	downloads        map[string]*downloadItem
	downloadsMutex   sync.Mutex
	downloadList     *DownloadList
	detailsContainer *fyne.Container
	selectedDownload *downloadDetails
	detailsCard      *widget.Card
	downloadSpeeds   map[string]float64
	globalSpeed      float64
	globalSpeedLabel *widget.Label
	lastSpeedUpdate  time.Time
	detailsPanel     *DetailsPanel
	app              fyne.App
	sideMenu         *fyne.Container
	menuButton       *widget.Button
	isMenuExpanded   bool
	db               *database.Database
}

func NewUI(d *downloader.Downloader, db *database.Database) *UI {
	a := app.New()
	ui := &UI{
		app:             a,
		downloader:      d,
		downloads:       make(map[string]*downloadItem),
		downloadSpeeds:  make(map[string]float64),
		lastSpeedUpdate: time.Now(),
		isMenuExpanded:  false,
		db:              db,
	}
	d.SetProgressCallback(ui.updateProgress)
	ui.downloadList = NewDownloadList(ui, db)
	ui.detailsPanel = NewDetailsPanel(ui, db)

	ui.sideMenu = ui.createSideMenu()

	return ui
}

func (u *UI) createSideMenu() *fyne.Container {
	filterItems := []struct {
		icon fyne.Resource
		text string
	}{
		{theme.ListIcon(), T("all")},
		{theme.DownloadIcon(), T("inProgress")},
		{theme.ConfirmIcon(), T("completed")},
		{theme.DeleteIcon(), T("deleted")},
		{theme.ErrorIcon(), T("errors")},
	}

	var filterButtons []fyne.CanvasObject
	for _, item := range filterItems {
		icon := widget.NewIcon(item.icon)
		label := widget.NewLabelWithStyle(item.text, fyne.TextAlignLeading, fyne.TextStyle{})
		label.Wrapping = fyne.TextWrapOff

		buttonContent := container.NewHBox(
			icon,
			label,
		)

		button := widget.NewButton("", func() {
			u.filterDownloads("", item.text)
		})
		button.Importance = widget.LowImportance

		customButton := container.NewStack(button, buttonContent)
		customButton.Resize(fyne.NewSize(400, 40))

		filterButtons = append(filterButtons, customButton)
	}

	filterContainer := container.NewVBox(filterButtons...)

	backgroundRect := canvas.NewRectangle(theme.BackgroundColor())

	boder := container.NewStack(
		backgroundRect,
		container.NewBorder(
			container.NewVBox(
				widget.NewLabelWithStyle(T("filters"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewSeparator(),
			),
			nil, nil, nil,
			container.NewStack(container.NewVScroll(filterContainer)),
		))

	return boder
}

// Ajoutez cette nouvelle méthode
func (u *UI) updateSideMenuSize() {
	if u.window != nil && u.sideMenu != nil {
		menuWidth := float32(220)
		u.sideMenu.Resize(fyne.NewSize(menuWidth, u.window.Canvas().Size().Height))

		// Mettre à jour la taille du rectangle d'arrière-plan
		if backgroundRect, ok := u.sideMenu.Objects[0].(*canvas.Rectangle); ok {
			backgroundRect.Resize(fyne.NewSize(menuWidth, u.window.Canvas().Size().Height))
		}
	}
}

func (u *UI) Start() {
	u.app.Settings().SetTheme(&myTheme{})
	u.window = u.app.NewWindow(T("windowTitle"))

	u.downloads = make(map[string]*downloadItem)
	u.downloadList = NewDownloadList(u, u.db)

	u.loadExistingDownloads()

	toolbar := NewToolbar(u)

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder(T("searchPlaceholder"))

	searchIcon := widget.NewIcon(theme.SearchIcon())
	searchBar := container.NewBorder(nil, nil, searchIcon, nil, searchEntry)

	u.isMenuExpanded = true

	u.globalSpeedLabel = widget.NewLabel(fmt.Sprintf(T("globalSpeed"), "0 B/s"))

	mainContent := container.NewVSplit(
		container.NewPadded(container.NewVScroll(u.downloadList.container)),
		u.detailsPanel.card,
	)
	mainContent.SetOffset(1)

	menuWidth := float32(220)
	u.sideMenu.Resize(fyne.NewSize(menuWidth, u.window.Canvas().Size().Height))
	menuContainer := container.NewHBox(u.sideMenu, widget.NewSeparator())

	overlay := container.NewStack(mainContent, menuContainer)

	u.menuButton = widget.NewButtonWithIcon("", theme.MenuIcon(), u.toggleMenu)

	topBar := container.NewVBox(
		container.NewBorder(nil, nil, u.menuButton, nil, toolbar.ToolbarObject()),
		searchBar,
		u.globalSpeedLabel,
	)

	content := container.NewBorder(
		topBar,
		nil, nil, nil,
		overlay,
	)

	u.window.SetContent(content)
	u.window.Resize(fyne.NewSize(1000, 600))

	u.updateSideMenuSize() // Ajoutez cette ligne après avoir créé la fenêtre

	u.window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		if key.Name == fyne.KeyF11 {
			u.window.SetFullScreen(!u.window.FullScreen())
		}
		u.updateSideMenuSize() // Mettez à jour la taille du menu latéral ici aussi
	})

	u.window.ShowAndRun()

	u.downloadList.Initialize()

	// Charger les paramètres depuis la base de données
	lang, err := u.db.GetSetting("language")
	if err != nil {
		log.Printf("Erreur lors du chargement de la langue : %v", err)
	} else if lang != "" {
		SetLanguage(language.MustParse(lang))
	}

	downloadDir, err := u.db.GetSetting("download_dir")
	if err != nil {
		log.Printf("Erreur lors du chargement du dossier de téléchargement : %v", err)
	} else if downloadDir != "" {
		u.downloader.DownloadDir = downloadDir
	}

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		for range ticker.C {
			u.updateDynamicElements()
		}
	}()
}

func (u *UI) updateDynamicElements() {
	u.downloadsMutex.Lock()
	defer u.downloadsMutex.Unlock()

	for _, item := range u.downloads {
		item.progressBar.Refresh()
		item.speedLabel.Refresh()
	}
	u.globalSpeedLabel.Refresh()
}

func (u *UI) getClipboardContent() string {
	return u.app.Driver().AllWindows()[0].Clipboard().Content()
}

func (u *UI) toggleMenu() {
	if u.sideMenu == nil {
		return
	}
	if u.isMenuExpanded {
		u.sideMenu.Hide()
		u.menuButton.SetIcon(theme.MenuIcon())
	} else {
		u.sideMenu.Show()
		u.menuButton.SetIcon(theme.NavigateBackIcon())
	}
	u.isMenuExpanded = !u.isMenuExpanded

	u.window.Content().Refresh()
}

func (u *UI) loadExistingDownloads() {
	downloads, err := u.db.GetAllDownloads()
	if err != nil {
		u.showError(T("errorTitle"), T("errorLoadingDownloads"))
		return
	}

	for _, download := range downloads {
		u.downloadList.addDownloadProgressToList(download.URL, download.Status)
	}
}

func (u *UI) updateDownloadStatus(url, status string) {
	u.downloadsMutex.Lock()
	defer u.downloadsMutex.Unlock()

	if item, exists := u.downloads[url]; exists {
		item.status = status
		if status == "completed" {
			item.progressBar.SetValue(1)
		}
	}
}

func (u *UI) updateProgress(url string, progress float64) {
	u.downloadList.updateProgress(url, progress)

	if u.selectedDownload != nil && u.selectedDownload.URL == url {
		u.detailsPanel.updateProgress()
	}
}

func (u *UI) updatePauseResumeButton(url string) {
	if item, exists := u.downloads[url]; exists {
		switch item.status {
		case "paused":
			item.pauseResumeButton.SetIcon(theme.MediaPlayIcon())
			item.pauseResumeButton.Show()
		case "downloading", "pending":
			item.pauseResumeButton.SetIcon(theme.MediaPauseIcon())
			item.pauseResumeButton.Show()
		case "completed":
			item.pauseResumeButton.Hide()
		}
	}
}

func formatSpeed(bytesPerSecond float64) string {
	if bytesPerSecond < 0 || bytesPerSecond == float64(^uint(0)>>1) {
		return "0 B/s"
	}
	units := []string{"B/s", "KB/s", "MB/s", "GB/s"}
	unitIndex := 0
	for bytesPerSecond >= 1024 && unitIndex < len(units)-1 {
		bytesPerSecond /= 1024
		unitIndex++
	}
	return fmt.Sprintf("%.2f %s", bytesPerSecond, units[unitIndex])
}

func (u *UI) showError(_ string, message string) {
	dialog.ShowError(errors.New(message), u.window)
}

func (u *UI) showInfo(title, message string) {
	dialog.ShowInformation(title, message, u.window)
}

func (u *UI) downloadMultiple(urlsText string) {
	urls := strings.Split(urlsText, "\n")
	validUrls := []string{}

	for _, urlStr := range urls {
		urlStr = strings.TrimSpace(urlStr)
		if urlStr == "" {
			continue
		}
		_, err := url.ParseRequestURI(urlStr)
		if err != nil {
			u.showError(T("errorTitle"), fmt.Sprintf(T("invalidURL"), urlStr))
		} else {
			validUrls = append(validUrls, urlStr)
			u.downloadList.addDownloadProgressToList(urlStr, "pending")
		}
	}

	if len(validUrls) == 0 {
		u.showError(T("errorTitle"), T("noValidURL"))
		return
	}

	errors := u.downloader.DownloadMultiple(validUrls)

	successCount := 0
	for i, err := range errors {
		if err == nil {
			successCount++
			u.downloadList.updateDownloadStatus(validUrls[i], "completed")
		} else {
			u.showError(T("downloadErrorTitle"), err.Error())
			u.downloadList.updateDownloadStatus(validUrls[i], "failed")
		}
	}

	u.showInfo(T("downloadsCompleted"), fmt.Sprintf(T("downloadsCompletedMessage"), successCount, len(validUrls)))
}

func (u *UI) updateGlobalSpeed() {
	now := time.Now()
	if now.Sub(u.lastSpeedUpdate) >= speedUpdateInterval {
		var totalSpeed float64
		for _, speed := range u.downloadList.downloadSpeeds {
			totalSpeed += speed
		}
		u.globalSpeed = totalSpeed
		if u.globalSpeedLabel != nil {
			u.globalSpeedLabel.SetText(fmt.Sprintf(T("globalSpeed"), formatSpeed(totalSpeed)))
		}
		u.lastSpeedUpdate = now
	}
}

func (u *UI) filterDownloads(searchTerm, filter string) {
	u.downloadList.filterDownloads(searchTerm, filter)
}

func (u *UI) showSettingsDialog() {
	languageSelect := widget.NewSelect([]string{T("english"), T("french")}, func(selected string) {
		switch selected {
		case T("english"):
			SetLanguage(language.English)
		case T("french"):
			SetLanguage(language.French)
		}
		u.refreshUI()
	})
	if currentLang == language.English {
		languageSelect.SetSelected(T("english"))
	} else {
		languageSelect.SetSelected(T("french"))
	}

	destinationEntry := widget.NewEntry()
	destinationEntry.SetText(u.downloader.DownloadDir)
	destinationButton := widget.NewButton(T("choose"), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				destinationEntry.SetText(uri.Path())
			}
		}, u.window)
	})

	chunksEntry := widget.NewEntry()
	chunksEntry.SetText(fmt.Sprintf("%d", u.downloader.MaxChunks))

	content := container.NewVBox(
		widget.NewLabel(T("language")),
		languageSelect,
		widget.NewLabel(T("destinationFolder")),
		container.NewBorder(nil, nil, nil, destinationButton, destinationEntry),
		widget.NewLabel(T("numberOfChunks")),
		chunksEntry,
	)

	dialog.ShowCustomConfirm(T("settings"), T("save"), T("cancel"), content, func(save bool) {
		if save {
			u.downloader.DownloadDir = destinationEntry.Text
			err := u.db.SetSetting("download_dir", u.downloader.DownloadDir)
			if err != nil {
				log.Printf("Erreur lors de l'enregistrement du dossier de téléchargement : %v", err)
				u.showError(T("errorTitle"), T("errorSavingSettings"))
				return
			}

			maxChunks, err := strconv.Atoi(chunksEntry.Text)
			if err == nil && maxChunks > 0 {
				u.downloader.MaxChunks = maxChunks
			}

			var langCode string
			switch languageSelect.Selected {
			case T("french"):
				langCode = "fr"
			default:
				langCode = "en"
			}
			err = u.db.SetSetting("language", langCode)
			if err != nil {
				log.Printf("Erreur lors de l'enregistrement de la langue : %v", err)
				u.showError(T("errorTitle"), T("errorSavingSettings"))
				return
			}

			// Appliquer immédiatement le changement de langue
			SetLanguage(language.MustParse(langCode))

			u.refreshUI()
			u.showInfo(T("settingsSaved"), T("settingsSavedMessage"))
		}
	}, u.window)
}

func (u *UI) refreshUI() {
	u.window.SetTitle(T("windowTitle"))
	u.globalSpeedLabel.SetText(fmt.Sprintf(T("globalSpeed"), formatSpeed(u.globalSpeed)))

	u.updateFilterTexts()

	u.updateSettingsButtonText()

	u.updateSearchPlaceholder()

	u.window.Content().Refresh()
}

func (u *UI) updateFilterTexts() {
	filterItems := []struct {
		icon fyne.Resource
		text string
	}{
		{theme.ListIcon(), T("all")},
		{theme.DownloadIcon(), T("inProgress")},
		{theme.ConfirmIcon(), T("completed")},
		{theme.DeleteIcon(), T("deleted")},
		{theme.ErrorIcon(), T("errors")},
	}

	if u.sideMenu == nil || len(u.sideMenu.Objects) == 0 {
		return
	}

	scrollContainer, ok := u.sideMenu.Objects[len(u.sideMenu.Objects)-1].(*container.Scroll)
	if !ok {
		return
	}

	vbox, ok := scrollContainer.Content.(*fyne.Container)
	if !ok {
		return
	}

	for i, item := range filterItems {
		if i < len(vbox.Objects) {
			if button, ok := vbox.Objects[i].(*fyne.Container); ok {
				if label, ok := button.Objects[1].(*widget.Label); ok {
					label.SetText(item.text)
				}
			}
		}
	}
}

func (u *UI) updateSettingsButtonText() {
	if content, ok := u.window.Content().(*fyne.Container); ok {
		if topBar, ok := content.Objects[0].(*fyne.Container); ok {
			if toolbar, ok := topBar.Objects[0].(*fyne.Container); ok {
				for _, obj := range toolbar.Objects {
					if settingsButton, ok := obj.(*widget.Button); ok {
						if settingsButton.Icon == theme.SettingsIcon() {
							settingsButton.SetText(T("settings"))
							break
						}
					}
				}
			}
		}
	}
}

func (u *UI) updateSearchPlaceholder() {
	if content, ok := u.window.Content().(*fyne.Container); ok {
		if topBar, ok := content.Objects[0].(*fyne.Container); ok {
			if searchBar, ok := topBar.Objects[1].(*fyne.Container); ok {
				if searchEntry, ok := searchBar.Objects[1].(*widget.Entry); ok {
					searchEntry.SetPlaceHolder(T("searchPlaceholder"))
				}
			}
		}
	}
}
