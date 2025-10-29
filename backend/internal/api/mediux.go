package api

import (
	"os"
	"path"
	"time"
)

var MediuxThumbsTempImageFolder string
var MediuxFullTempImageFolder string
var MediuxBaseURL string = "https://images.mediux.io"

func init() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "/config"
	}
	MediuxThumbsTempImageFolder = path.Join(configPath, "temp-images", "mediux", "thumbs")
	MediuxFullTempImageFolder = path.Join(configPath, "temp-images", "mediux", "full")
}

type MediuxResponse struct {
	Data struct {
		Movie *MediuxMovieByID `json:"movies_by_id,omitempty"`
		Show  *MediuxShowByID  `json:"shows_by_id,omitempty"`
	} `json:"data"`
}

type MediuxShowSetResponse struct {
	Data struct {
		ShowSetID MediuxShowSetByID `json:"show_sets_by_id"`
	} `json:"data"`
}

type MediuxMovieSetResponse struct {
	Data struct {
		MovieSetID MediuxMovieSetByID `json:"movie_sets_by_id"`
	} `json:"data"`
}

type MediuxCollectionSetResponse struct {
	Data struct {
		CollectionSetID MediuxCollectionSetByID `json:"collection_sets_by_id"`
	} `json:"data"`
}

type MediuxMovieSetByID struct {
	ID          string            `json:"id"`
	SetTitle    string            `json:"set_title,omitempty"`
	UserCreated MediuxUserCreated `json:"user_created"`
	DateCreated time.Time         `json:"date_created"`
	DateUpdated time.Time         `json:"date_updated"`
	Movie       MediuxMovieByID   `json:"movie_id"`
}

type MediuxCollectionSetByID struct {
	ID          string                  `json:"id"`
	SetTitle    string                  `json:"set_title,omitempty"`
	UserCreated MediuxUserCreated       `json:"user_created"`
	DateCreated time.Time               `json:"date_created"`
	DateUpdated time.Time               `json:"date_updated"`
	Collection  MediuxMovieCollectionID `json:"collection_id"`
}

type MediuxShowSetByID struct {
	ID          string            `json:"id"`
	SetTitle    string            `json:"set_title,omitempty"`
	UserCreated MediuxUserCreated `json:"user_created"`
	DateCreated time.Time         `json:"date_created"`
	DateUpdated time.Time         `json:"date_updated"`
	Show        MediuxShowByID    `json:"show_id"`
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
	CollectionSet MediuxSetInfo `json:"collection_set"`
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

type MediuxUserAllSetsResponse struct {
	Data struct {
		ShowSets       []MediuxUserShowSet       `json:"show_sets"`
		MovieSets      []MediuxUserMovieSet      `json:"movie_sets"`
		CollectionSets []MediuxUserCollectionSet `json:"collection_sets"`
		Boxsets        []MediuxUserBoxset        `json:"boxsets"`
	} `json:"data"`
}

type MediuxUserShowSet struct {
	ID            string                   `json:"id"`
	UserCreated   MediuxUserCreated        `json:"user_created"`
	SetTitle      string                   `json:"set_title"`
	DateCreated   time.Time                `json:"date_created"`
	DateUpdated   time.Time                `json:"date_updated"`
	ShowID        MediuxUserShow           `json:"show_id"`
	ShowPoster    []MediuxUserImage        `json:"show_poster"`
	ShowBackdrop  []MediuxUserImage        `json:"show_backdrop"`
	SeasonPosters []MediuxUserSeasonPoster `json:"season_posters"`
	Titlecards    []MediuxUserTitlecard    `json:"titlecards"`
}

type MediuxUserMovieSet struct {
	ID            string            `json:"id"`
	UserCreated   MediuxUserCreated `json:"user_created"`
	SetTitle      string            `json:"set_title"`
	DateCreated   time.Time         `json:"date_created"`
	DateUpdated   time.Time         `json:"date_updated"`
	MovieID       MediuxUserMovie   `json:"movie_id"`
	MoviePoster   []MediuxUserImage `json:"movie_poster"`
	MovieBackdrop []MediuxUserImage `json:"movie_backdrop"`
}

type MediuxUserCollectionSet struct {
	ID             string                      `json:"id"`
	UserCreated    MediuxUserCreated           `json:"user_created"`
	SetTitle       string                      `json:"set_title"`
	DateCreated    time.Time                   `json:"date_created"`
	DateUpdated    time.Time                   `json:"date_updated"`
	MoviePosters   []MediuxUserCollectionMovie `json:"movie_posters"`
	MovieBackdrops []MediuxUserCollectionMovie `json:"movie_backdrops"`
}

type MediuxUserBoxset struct {
	ID             string                    `json:"id"`
	UserCreated    MediuxUserCreated         `json:"user_created"`
	BoxsetTitle    string                    `json:"boxset_title"`
	DateCreated    time.Time                 `json:"date_created"`
	DateUpdated    time.Time                 `json:"date_updated"`
	MovieSets      []MediuxUserMovieSet      `json:"movie_sets"`
	ShowSets       []MediuxUserShowSet       `json:"show_sets"`
	CollectionSets []MediuxUserCollectionSet `json:"collection_sets"`
}

type MediuxUserShow struct {
	ID           string    `json:"id"`
	DateUpdated  time.Time `json:"date_updated"`
	Status       string    `json:"status"`
	Title        string    `json:"title"`
	Tagline      string    `json:"tagline"`
	FirstAirDate string    `json:"first_air_date"`
	TvdbID       string    `json:"tvdb_id"`
	ImdbID       string    `json:"imdb_id"`
	TraktID      string    `json:"trakt_id"`
	Slug         string    `json:"slug"`
}

type MediuxUserMovie struct {
	ID          string    `json:"id"`
	DateUpdated time.Time `json:"date_updated"`
	Status      string    `json:"status"`
	Title       string    `json:"title"`
	Tagline     string    `json:"tagline"`
	ReleaseDate string    `json:"release_date"`
	TvdbID      string    `json:"tvdb_id"`
	ImdbID      string    `json:"imdb_id"`
	TraktID     string    `json:"trakt_id"`
	Slug        string    `json:"slug"`
}

type MediuxUserImage struct {
	ID         string    `json:"id"`
	ModifiedOn time.Time `json:"modified_on"`
	Filesize   string    `json:"filesize"`
}

type MediuxUserSeasonPoster struct {
	ID         string    `json:"id"`
	ModifiedOn time.Time `json:"modified_on"`
	Filesize   string    `json:"filesize"`
	Season     struct {
		SeasonNumber int `json:"season_number"`
	} `json:"season"`
}

type MediuxUserTitlecard struct {
	ID         string    `json:"id"`
	ModifiedOn time.Time `json:"modified_on"`
	Filesize   string    `json:"filesize"`
	Episode    struct {
		EpisodeTitle  string `json:"episode_title"`
		EpisodeNumber int    `json:"episode_number"`
		SeasonID      struct {
			SeasonNumber int `json:"season_number"`
		} `json:"season_id"`
	} `json:"episode"`
}

type MediuxUserCollectionMovie struct {
	ID         string    `json:"id"`
	ModifiedOn time.Time `json:"modified_on"`
	Filesize   string    `json:"filesize"`
	Movie      struct {
		ID          string    `json:"id"`
		DateUpdated time.Time `json:"date_updated"`
		Status      string    `json:"status"`
		Title       string    `json:"title"`
		Tagline     string    `json:"tagline"`
		ReleaseDate string    `json:"release_date"`
		TvdbID      string    `json:"tvdb_id"`
		ImdbID      string    `json:"imdb_id"`
		TraktID     string    `json:"trakt_id"`
		Slug        string    `json:"slug"`
	} `json:"movie"`
}
