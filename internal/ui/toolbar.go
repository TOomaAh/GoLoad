package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Toolbar struct {
	ui        *UI
	container *widget.Toolbar
}

func NewToolbar(ui *UI) *Toolbar {
	t := &Toolbar{ui: ui}
	t.container = widget.NewToolbar(
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.ContentAddIcon(), t.showAddDownloadDialog),
		widget.NewToolbarAction(theme.MediaPlayIcon(), t.resumeDownloads),
		widget.NewToolbarAction(theme.SettingsIcon(), t.showSettingsDialog),
		widget.NewToolbarSpacer(),
	)
	return t
}

func (t *Toolbar) ToolbarObject() fyne.CanvasObject {
	return t.container
}

func (t *Toolbar) showAddDownloadDialog() {
	showAddDownloadDialog(t.ui)
}

func (t *Toolbar) resumeDownloads() {
	resumeDownloads(t.ui)
}

func (t *Toolbar) showSettingsDialog() {
	t.ui.showSettingsDialog()
}
