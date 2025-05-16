package modals

type Database_SavedSet struct {
	ID            string         `json:"ID"`
	MediaItem     MediaItem      `json:"MediaItem"`
	MediaItemJSON string         `json:"MediaItemJSON"`
	Sets          []Database_Set `json:"Sets"`
}

type Database_Set struct {
	ID            string    `json:"ID"`
	MediaItemID   string    `json:"MediaItemID"`
	Set           PosterSet `json:"Set"`
	SetJSON       string    `json:"SetJSON"`
	SelectedTypes []string  `json:"SelectedTypes"`
	AutoDownload  bool      `json:"AutoDownload"`
	LastUpdate    string    `json:"LastUpdate,omitempty"`
	ToDelete      bool      `json:"ToDelete,omitempty"`
}
