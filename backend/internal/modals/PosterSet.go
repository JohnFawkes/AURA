package modals

import "time"

type MediuxResponse struct {
	Data struct {
		Movie *MediuxMovieByID `json:"movies_by_id,omitempty"`
		Show  *MediuxShowByID  `json:"shows_by_id,omitempty"`
	} `json:"data"`
}

type MediuxShowSetResponse struct {
	Data struct {
		ShowSetID MediuxShowSetByID `json:"show_sets_by_id,omitempty"`
	} `json:"data"`
}

type MediuxMovieSetResponse struct {
	Data struct {
		MovieSetID MediuxMovieSetByID `json:"movie_sets_by_id,omitempty"`
	} `json:"data"`
}

type MediuxCollectionSetResponse struct {
	Data struct {
		CollectionSetID MediuxCollectionSetByID `json:"collection_sets_by_id,omitempty"`
	} `json:"data"`
}

type MediuxMovieSetByID struct {
	ID          string            `json:"id"`
	SetTitle    string            `json:"set_title,omitempty"`
	UserCreated MediuxUserCreated `json:"user_created"`
	DateCreated time.Time         `json:"date_created"`
	DateUpdated time.Time         `json:"date_updated"`
	Movie       MediuxMovieByID   `json:"movie_id,omitempty"`
}

type MediuxCollectionSetByID struct {
	ID          string                  `json:"id"`
	SetTitle    string                  `json:"set_title,omitempty"`
	UserCreated MediuxUserCreated       `json:"user_created"`
	DateCreated time.Time               `json:"date_created"`
	DateUpdated time.Time               `json:"date_updated"`
	Collection  MediuxMovieCollectionID `json:"collection_id,omitempty"`
}

type MediuxShowSetByID struct {
	ID          string            `json:"id"`
	SetTitle    string            `json:"set_title,omitempty"`
	UserCreated MediuxUserCreated `json:"user_created"`
	DateCreated time.Time         `json:"date_created"`
	DateUpdated time.Time         `json:"date_updated"`
	Show        MediuxShowByID    `json:"show_id,omitempty"`
}

type MediuxMovieByID struct {
	ID           string                       `json:"id"`
	DateUpdated  time.Time                    `json:"date_updated"`
	Status       string                       `json:"status"`
	Title        string                       `json:"title"`
	Tagline      string                       `json:"tagline"`
	ReleaseDate  string                       `json:"release_date"`
	TvdbID       string                       `json:"tvdb_id"`
	ImdbID       string                       `json:"imdb_id"`
	TraktID      string                       `json:"trakt_id"`
	Slug         string                       `json:"slug"`
	CollectionID *MediuxMovieCollectionID     `json:"collection_id,omitempty"`
	Posters      []MediuxMoviePosterSetImages `json:"posters,omitempty"`
	Backdrops    []MediuxMoviePosterSetImages `json:"backdrops,omitempty"`
}

type MediuxShowByID struct {
	ID           string                      `json:"id"`
	DateUpdated  time.Time                   `json:"date_updated"`
	Status       string                      `json:"status"`
	Title        string                      `json:"title"`
	Tagline      string                      `json:"tagline"`
	FirstAirDate string                      `json:"first_air_date"`
	TvdbID       string                      `json:"tvdb_id"`
	ImdbID       string                      `json:"imdb_id"`
	TraktID      string                      `json:"trakt_id"`
	Slug         string                      `json:"slug"`
	Posters      []MediuxShowPosterSetImages `json:"posters,omitempty"`
	Backdrops    []MediuxShowPosterSetImages `json:"backdrops,omitempty"`
	Seasons      []MediuxShowSeasons         `json:"seasons,omitempty"`
}

type MediuxMovieCollectionID struct {
	ID             string                       `json:"id"`
	CollectionName string                       `json:"collection_name"`
	Movies         []MediuxMovieCollectionMovie `json:"movies,omitempty"`
}

type MediuxMovieCollectionMovie struct {
	ID          string                        `json:"id"`
	DateUpdated time.Time                     `json:"date_updated"`
	Status      string                        `json:"status"`
	Title       string                        `json:"title"`
	Tagline     string                        `json:"tagline"`
	ReleaseDate string                        `json:"release_date"`
	TvdbID      string                        `json:"tvdb_id"`
	ImdbID      string                        `json:"imdb_id"`
	TraktID     string                        `json:"trakt_id"`
	Slug        string                        `json:"slug"`
	Posters     []MediuxMovieCollectionImages `json:"posters,omitempty"`
	Backdrops   []MediuxMovieCollectionImages `json:"backdrops,omitempty"`
}

type MediuxMovieCollectionImages struct {
	ID            string        `json:"id"`
	ModifiedOn    time.Time     `json:"modified_on"`
	FileSize      string        `json:"filesize"`
	CollectionSet MediuxSetInfo `json:"collection_set,omitempty"`
}

type MediuxSetInfo struct {
	ID          string            `json:"id"`
	SetTitle    string            `json:"set_title,omitempty"`
	UserCreated MediuxUserCreated `json:"user_created"`
	DateCreated time.Time         `json:"date_created"`
	DateUpdated time.Time         `json:"date_updated"`
}

type MediuxMoviePosterSetImages struct {
	ID         string         `json:"id"`
	ModifiedOn time.Time      `json:"modified_on"`
	FileSize   string         `json:"filesize"`
	MovieSet   *MediuxSetInfo `json:"movie_set,omitempty"`
}

type MediuxShowPosterSetImages struct {
	ID         string         `json:"id"`
	ModifiedOn time.Time      `json:"modified_on"`
	FileSize   string         `json:"filesize"`
	ShowSet    *MediuxSetInfo `json:"show_set,omitempty"`
}

type MediuxShowSeasons struct {
	SeasonNumber int                         `json:"season_number"`
	Posters      []MediuxShowPosterSetImages `json:"posters,omitempty"`
	Episodes     []MediuxShowEpisodes        `json:"episodes,omitempty"`
}

type MediuxShowEpisodes struct {
	EpisodeTitle  string                      `json:"episode_title,omitempty"`
	EpisodeNumber int                         `json:"episode_number,omitempty"`
	Season        *MediuxShowSeasons          `json:"season_id,omitempty"`
	Titlecards    []MediuxShowPosterSetImages `json:"titlecards,omitempty"`
}

type MediuxShowEpisodeSeason struct {
	SeasonNumber int `json:"season_number,omitempty"`
}

type MediuxUserCreated struct {
	Username string `json:"username"`
}

//////////////////////

type PosterSet struct {
	ID             string       `json:"ID"`
	Title          string       `json:"Title"`
	Type           string       `json:"Type"`
	User           SetUser      `json:"User"`
	DateCreated    time.Time    `json:"DateCreated"`
	DateUpdated    time.Time    `json:"DateUpdated"`
	Poster         *PosterFile  `json:"Poster,omitempty"`
	OtherPosters   []PosterFile `json:"OtherPosters,omitempty"` // User for movie collections
	Backdrop       *PosterFile  `json:"Backdrop,omitempty"`
	OtherBackdrops []PosterFile `json:"OtherBackdrops,omitempty"` // User for movie collections
	SeasonPosters  []PosterFile `json:"SeasonPosters,omitempty"`
	TitleCards     []PosterFile `json:"TitleCards,omitempty"`
	Status         string       `json:"Status,omitempty"`
}

type PosterFile struct {
	ID       string             `json:"ID"`
	Type     string             `json:"Type"`
	Modified time.Time          `json:"Modified"`
	FileSize int64              `json:"FileSize"`
	Movie    *PosterFileMovie   `json:"Movie,omitempty"`
	Show     *PosterFileShow    `json:"Show,omitempty"`
	Season   *PosterFileSeason  `json:"Season,omitempty"`
	Episode  *PosterFileEpisode `json:"Episode,omitempty"`
}

type PosterFileShow struct {
	ID             string    `json:"ID"`
	Title          string    `json:"Title"`
	RatingKey      string    `json:"RatingKey,omitempty"`
	LibrarySection string    `json:"LibrarySection,omitempty"`
	MediaItem      MediaItem `json:"MediaItem"`
}
type PosterFileMovie struct {
	ID             string    `json:"ID"`
	Title          string    `json:"Title"`
	Status         string    `json:"Status,omitempty"`
	Tagline        string    `json:"Tagline,omitempty"`
	Slug           string    `json:"Slug,omitempty"`
	DateUpdated    time.Time `json:"DateUpdated"`
	TvdbID         string    `json:"TvdbID,omitempty"`
	ImdbID         string    `json:"ImdbID,omitempty"`
	TraktID        string    `json:"TraktID,omitempty"`
	ReleaseDate    string    `json:"ReleaseDate,omitempty"`
	RatingKey      string    `json:"RatingKey,omitempty"`
	LibrarySection string    `json:"LibrarySection,omitempty"`
	MediaItem      MediaItem `json:"MediaItem"`
}

type PosterFileSeason struct {
	Number int `json:"Number"`
}

type PosterFileEpisode struct {
	Title         string `json:"Title"`
	EpisodeNumber int    `json:"EpisodeNumber"`
	SeasonNumber  int    `json:"SeasonNumber"`
}

type SetUser struct {
	Name string `json:"Name"`
}
