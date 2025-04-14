package modals

type LibrarySection struct {
	ID         string      `json:"ID"`
	Type       string      `json:"Type"` // "movie" or "show"
	Title      string      `json:"Title"`
	MediaItems []MediaItem `json:"MediaItems"`
}

type MediaItem struct {
	RatingKey      string  `json:"RatingKey"`
	Type           string  `json:"Type"` // "movie" or "series"
	Title          string  `json:"Title"`
	Year           int     `json:"Year"`
	Thumb          string  `json:"Thumb"`
	AudienceRating float64 `json:"AudienceRating"`
	UserRating     float64 `json:"UserRating"`
	ContentRating  string  `json:"ContentRating"`
	Summary        string  `json:"Summary"`
	UpdatedAt      int64   `json:"UpdatedAt"`
	Guids          []Guid  `json:"Guids"`

	Movie  *PlexMovie  `json:"Movie,omitempty"`
	Series *PlexSeries `json:"Series,omitempty"`
}

type Guid struct {
	Provider string `json:"Provider"`
	ID       string `json:"ID"`
}

type PlexMovie struct {
	File PlexFile `json:"File"`
}

type PlexSeries struct {
	Seasons      []PlexSeason `json:"Seasons"`
	SeasonCount  int          `json:"SeasonCount"`
	EpisodeCount int          `json:"EpisodeCount"`
	Location     string       `json:"Location"`
}

type PlexSeason struct {
	RatingKey    string        `json:"RatingKey"`
	SeasonNumber int           `json:"SeasonNumber"`
	Title        string        `json:"Title"`
	Episodes     []PlexEpisode `json:"Episodes"`
}

type PlexEpisode struct {
	RatingKey     string   `json:"RatingKey"`
	Title         string   `json:"Title"`
	SeasonNumber  int      `json:"SeasonNumber"`
	EpisodeNumber int      `json:"EpisodeNumber"`
	File          PlexFile `json:"File"`
}

type PlexFile struct {
	Path     string `json:"Path"`
	Size     int64  `json:"Size"`
	Duration int64  `json:"Duration"`
}
