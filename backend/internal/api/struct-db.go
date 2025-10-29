package api

// PosterSetDetail groups poster set details per media item.
type DBPosterSetDetail struct {
	PosterSetID    string    `json:"PosterSetID"`
	PosterSet      PosterSet `json:"PosterSet"`
	PosterSetJSON  string    `json:"PosterSetJSON"`
	LastDownloaded string    `json:"LastDownloaded"`
	SelectedTypes  []string  `json:"SelectedTypes"`
	AutoDownload   bool      `json:"AutoDownload"`
	ToDelete       bool      `json:"ToDelete"` // Flag to indicate if the poster set should be deleted (Not used in DB)
}

// MediaItemWithPosterSets groups a media item with its poster sets.
type DBMediaItemWithPosterSets struct {
	TMDB_ID       string              `json:"TMDB_ID"`
	LibraryTitle  string              `json:"LibraryTitle"`
	MediaItem     MediaItem           `json:"MediaItem"`
	MediaItemJSON string              `json:"MediaItemJSON"`
	PosterSets    []DBPosterSetDetail `json:"PosterSets"`
}
