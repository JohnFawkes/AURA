package modals

type DBSavedItem struct {
	MediaItemID    string    `json:"MediaItemID"`    // ID of the Item from the Media Server
	MediaItem      MediaItem `json:"MediaItem"`      // MediaItem
	MediaItemJSON  string    `json:"MediaItemJSON"`  // JSON of the MediaItem
	PosterSetID    string    `json:"PosterSetID"`    // ID of the PosterSet
	PosterSet      PosterSet `json:"PosterSet"`      // PosterSet
	PosterSetJSON  string    `json:"PosterSetJSON"`  // JSON of the PosterSet
	LastDownloaded string    `json:"LastDownloaded"` // Last Downloaded Date
	SelectedTypes  []string  `json:"SelectedTypes"`  // Types of Posters/Backdrops/Season Posters/Titlecards
	AutoDownload   bool      `json:"AutoDownload"`   // Auto Download
}

// PosterSetDetail groups poster set details per media item.
type DBPosterSetDetail struct {
	PosterSetID    string    `json:"PosterSetID"`
	PosterSet      PosterSet `json:"PosterSet"`
	PosterSetJSON  string    `json:"PosterSetJSON"`
	LastDownloaded string    `json:"LastDownloaded"`
	SelectedTypes  []string  `json:"SelectedTypes"`
	AutoDownload   bool      `json:"AutoDownload"`
	ToDelete       bool      `json:"ToDelete"` // Flag to indicate if the poster set should be deleted
}

// MediaItemWithPosterSets groups a media item with its poster sets.
type DBMediaItemWithPosterSets struct {
	MediaItemID   string              `json:"MediaItemID"`
	MediaItem     MediaItem           `json:"MediaItem"`
	MediaItemJSON string              `json:"MediaItemJSON"`
	PosterSets    []DBPosterSetDetail `json:"PosterSets"`
}
