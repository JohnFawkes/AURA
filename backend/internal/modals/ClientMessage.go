package modals

type ClientMessage struct {
	MediaItem     MediaItem `json:"MediaItem"`
	Set           PosterSet `json:"Set"`
	SelectedTypes []string  `json:"SelectedTypes"`
	AutoDownload  bool      `json:"AutoDownload"`
	LastUpdate    string    `json:"LastUpdate,omitempty"`
}
