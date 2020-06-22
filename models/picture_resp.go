package models

import "time"

type PicturesResponse struct {
	MediaItems    []*MediaItem `json:"mediaItems"`
	NextPageToken string       `json:"nextPageToken"`
}

type Video struct {
	CameraMake  string  `json:"cameraMake"`
	CameraModel string  `json:"cameraModel"`
	Fps         float64 `json:"fps"`
	Status      string  `json:"status"`
}

type MediaMetadata struct {
	CreationTime time.Time `json:"creationTime"`
	Width        string    `json:"width"`
	Height       string    `json:"height"`
	Video        *Video     `json:"video"`
}

type MediaItem struct {
	ID            string        `json:"id"`
	ProductURL    string        `json:"productUrl"`
	BaseURL       string        `json:"baseUrl"`
	MimeType      string        `json:"mimeType"`
	MediaMetadata *MediaMetadata `json:"mediaMetadata"`
	Filename      string        `json:"filename"`
	Ready		  chan bool `json:"-"`
}