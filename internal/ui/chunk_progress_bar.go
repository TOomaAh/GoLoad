package ui

import (
	"gestionnaire-telechargement/internal/downloader"

	"fyne.io/fyne/v2/widget"
)

type ChunkProgressBar struct {
	widget.ProgressBar
	chunks []downloader.ChunkInfo
}

func NewChunkProgressBar(chunks []downloader.ChunkInfo) *ChunkProgressBar {
	bar := &ChunkProgressBar{
		chunks: chunks,
	}
	bar.ExtendBaseWidget(bar)
	return bar
}

func (c *ChunkProgressBar) UpdateChunks(chunks []downloader.ChunkInfo) {
	c.chunks = chunks
	c.Refresh()
}
