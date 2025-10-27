package api

type LibrarySection struct {
	ID         string      `json:"ID"`
	Type       string      `json:"Type"` // "movie" or "show"
	Title      string      `json:"Title"`
	TotalSize  int         `json:"TotalSize"`
	MediaItems []MediaItem `json:"MediaItems"`
	Path       string      `json:"Path,omitempty"`
}

type MediaItem struct {
	TMDB_ID         string             `json:"TMDB_ID"`                 // TMDB ID of the media item
	LibraryTitle    string             `json:"LibraryTitle"`            // Title of the library/section the item belongs to
	RatingKey       string             `json:"RatingKey"`               // RatingKey is the internal ID from the media server
	Type            string             `json:"Type"`                    // "movie" or "show"
	Title           string             `json:"Title"`                   // Title of the media item
	Year            int                `json:"Year"`                    // Release year of the media item
	ExistInDatabase bool               `json:"ExistInDatabase"`         // Indicates if the item exists in the local database
	DBSavedSets     []PosterSetSummary `json:"DBSavedSets,omitempty"`   // Poster sets saved in the database for this item
	Thumb           string             `json:"Thumb,omitempty"`         // URL to the thumbnail image
	ContentRating   string             `json:"ContentRating,omitempty"` // Content rating (e.g., "PG-13")
	Summary         string             `json:"Summary,omitempty"`       // Summary or description of the media item
	UpdatedAt       int64              `json:"UpdatedAt,omitempty"`     // Last updated timestamp
	AddedAt         int64              `json:"AddedAt,omitempty"`       // Added timestamp
	ReleasedAt      int64              `json:"ReleasedAt,omitempty"`    // Release date timestamp
	Guids           []Guid             `json:"Guids,omitempty"`         // List of GUIDs from different providers
	Movie           *MediaItemMovie    `json:"Movie,omitempty"`         // Present if Type is "movie"; Contains file info
	Series          *MediaItemSeries   `json:"Series,omitempty"`        // Present if Type is "show"; Contains seasons and episodes info
}

type Guid struct {
	Provider string `json:"Provider"`
	ID       string `json:"ID"`
	Rating   string `json:"Rating"`
}

type MediaItemMovie struct {
	File MediaItemFile `json:"File"`
}

type MediaItemSeries struct {
	Seasons      []MediaItemSeason `json:"Seasons"`
	SeasonCount  int               `json:"SeasonCount"`
	EpisodeCount int               `json:"EpisodeCount"`
	Location     string            `json:"Location"`
}

type MediaItemSeason struct {
	RatingKey    string             `json:"RatingKey"`
	SeasonNumber int                `json:"SeasonNumber"`
	Title        string             `json:"Title"`
	Episodes     []MediaItemEpisode `json:"Episodes"`
}

type MediaItemEpisode struct {
	RatingKey     string        `json:"RatingKey"`
	Title         string        `json:"Title"`
	SeasonNumber  int           `json:"SeasonNumber"`
	EpisodeNumber int           `json:"EpisodeNumber"`
	File          MediaItemFile `json:"File"`
}

type MediaItemFile struct {
	Path     string `json:"Path"`
	Size     int64  `json:"Size"`
	Duration int64  `json:"Duration"`
}

type PosterSetSummary struct {
	PosterSetID   string
	PosterSetUser string
	SelectedTypes []string
}
