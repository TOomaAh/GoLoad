package models

import "gorm.io/gorm"

type Settings struct {
	gorm.Model
	DarkMode     bool   `json:"darkMode"`
	DownloadPath string `json:"downloadPath"`
	MaxDownloads int    `json:"maxDownloads"`
	MaxSpeed     int64  `json:"maxSpeed"`

	AutoStartDownloads        bool `json:"autoStartDownloads"`
	RetryAfterFailure         bool `json:"retryAfterFailure"`
	NotificationEndDownload   bool `json:"notificationEndDownload"`
	NotificationErrorDownload bool `json:"notificationErrorDownload"`
}
