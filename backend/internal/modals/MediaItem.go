package modals

type LibrarySection struct {
	ID         string      `json:"ID"`
	Type       string      `json:"Type"` // "movie" or "show"
	Title      string      `json:"Title"`
	TotalSize  int         `json:"TotalSize"`
	MediaItems []MediaItem `json:"MediaItems"`
}

type MediaItem struct {
	RatingKey      string  `json:"RatingKey"`
	Type           string  `json:"Type"` // "movie" or "series"
	LibraryTitle   string  `json:"LibraryTitle"`
	Title          string  `json:"Title"`
	Year           int     `json:"Year"`
	Thumb          string  `json:"Thumb,omitempty"`
	AudienceRating float64 `json:"AudienceRating,omitempty"`
	UserRating     float64 `json:"UserRating,omitempty"`
	ContentRating  string  `json:"ContentRating,omitempty"`
	Summary        string  `json:"Summary,omitempty"`
	UpdatedAt      int64   `json:"UpdatedAt,omitempty"`
	Guids          []Guid  `json:"Guids,omitempty"`

	Movie  *MediaItemMovie  `json:"Movie,omitempty"`
	Series *MediaItemSeries `json:"Series,omitempty"`
}

type Guid struct {
	Provider string `json:"Provider"`
	ID       string `json:"ID"`
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
