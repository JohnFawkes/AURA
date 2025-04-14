package modals

type ClientMessage struct {
	Plex          MediaItem `json:"Plex"`
	Set           PosterSet `json:"Set"`
	SelectedTypes []string  `json:"SelectedTypes"`
	AutoDownload  bool      `json:"AutoDownload"`
	LastUpdate    string    `json:"LastUpdate,omitempty"`
}
