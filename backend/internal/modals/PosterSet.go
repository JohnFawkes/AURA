package modals

import "time"

type MediuxResponse struct {
	Data struct {
		Movies *[]MediuxItem `json:"movies,omitempty"`
		Shows  *[]MediuxItem `json:"shows,omitempty"`
	} `json:"data"`
}

type MediuxFileItem struct {
	ID       string             `json:"id"`
	FileType string             `json:"file_type"`
	Movie    *MediuxFileMovie   `json:"movie,omitempty"`
	Season   *MediuxFileSeason  `json:"season,omitempty"`
	Episode  *MediuxFileEpisode `json:"episode,omitempty"`
}

type MediuxFileMovie struct {
	ID string `json:"id,omitempty"`
}

type MediuxFileSeason struct {
	SeasonNumber int `json:"season_number,omitempty"`
}

type MediuxFileEpisode struct {
	EpisodeTitle  string            `json:"episode_title,omitempty"`
	EpisodeNumber int               `json:"episode_number,omitempty"`
	Season        *MediuxFileSeason `json:"season_id,omitempty"`
}

type MediuxUserCreated struct {
	Username string `json:"username"`
}

type MediuxItem struct {
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	Status       string              `json:"status"`
	Tagline      string              `json:"tagline"`
	Slug         string              `json:"slug"`
	DateUpdated  time.Time           `json:"date_updated"`
	TvdbID       string              `json:"tvdb_id"`
	ImdbID       string              `json:"imdb_id"`
	TraktID      string              `json:"trakt_id"`
	ReleaseDate  string              `json:"release_date,omitempty"`
	FirstAirDate string              `json:"first_air_date,omitempty"`
	CollectionID *MediuxCollectionID `json:"collection_id,omitempty"`
	MovieSets    *[]MediuxPosterSet  `json:"movie_sets,omitempty"`
	ShowSets     *[]MediuxPosterSet  `json:"show_sets,omitempty"`
}

type MediuxCollectionID struct {
	ID             string                `json:"id"`
	CollectionName string                `json:"collection_name"`
	CollectionSets []MediuxCollectionSet `json:"collection_sets,omitempty"`
}

type MediuxCollectionSet struct {
	ID          string            `json:"id"`
	UserCreated MediuxUserCreated `json:"user_created"`
	DateCreated time.Time         `json:"date_created"`
	DateUpdated time.Time         `json:"date_updated"`
	Files       []MediuxFileItem  `json:"files"`
}

type MediuxPosterSet struct {
	ID          string            `json:"id"`
	UserCreated MediuxUserCreated `json:"user_created"`
	DateCreated time.Time         `json:"date_created"`
	DateUpdated time.Time         `json:"date_updated"`
	Files       []MediuxFileItem  `json:"files"`
}

type PosterSets struct {
	Type string      `json:"Type,omitempty"`
	Item PosterItem  `json:"Item,omitempty"`
	Sets []PosterSet `json:"Sets,omitempty"`
}

type PosterItem struct {
	ID           string    `json:"ID"`
	Title        string    `json:"Title"`
	Status       string    `json:"Status"`
	Tagline      string    `json:"Tagline"`
	Slug         string    `json:"Slug"`
	DateUpdated  time.Time `json:"DateUpdated,omitempty"`
	TvdbID       string    `json:"TvdbID"`
	ImdbID       string    `json:"ImdbID"`
	TraktID      string    `json:"TraktID"`
	FirstAirDate string    `json:"FirstAirDate,omitempty"`
	ReleaseDate  string    `json:"ReleaseDate,omitempty"`
}

type PosterSet struct {
	ID    string `json:"ID"`
	Title string `json:"Title"`
	Type  string `json:"Type"`
	User  struct {
		Name string `json:"Name"`
	} `json:"User"`
	DateCreated time.Time    `json:"DateCreated"`
	DateUpdated time.Time    `json:"DateUpdated"`
	Files       []PosterFile `json:"Files"`
}

type PosterFile struct {
	ID      string             `json:"ID"`
	Type    string             `json:"Type"`
	Movie   *PosterFileMovie   `json:"Movie,omitempty"`
	Season  *PosterFileSeason  `json:"Season,omitempty"`
	Episode *PosterFileEpisode `json:"Episode,omitempty"`
}

type PosterFileMovie struct {
	ID string `json:"ID,omitempty"`
}

type PosterFileSeason struct {
	Number int `json:"Number,omitempty"`
}

type PosterFileEpisode struct {
	Title         string `json:"Title,omitempty"`
	EpisodeNumber int    `json:"EpisodeNumber,omitempty"`
	SeasonNumber  int    `json:"SeasonNumber,omitempty"`
}

// type MediuxMovieSetsResponse struct {
// 	Data struct {
// 		Movies []struct {
// 			ID           string    `json:"id"`
// 			DateUpdated  time.Time `json:"date_updated"`
// 			Status       string    `json:"status"`
// 			Title        string    `json:"title"`
// 			ReleaseDate  string    `json:"release_date"`
// 			Tagline      string    `json:"tagline"`
// 			TvdbID       string    `json:"tvdb_id"`
// 			ImdbID       string    `json:"imdb_id"`
// 			TraktID      string    `json:"trakt_id"`
// 			Slug         string    `json:"slug"`
// 			CollectionID struct {
// 				ID             string `json:"id"`
// 				CollectionName string `json:"collection_name"`
// 				CollectionSets []struct {
// 					ID          string `json:"id"`
// 					UserCreated struct {
// 						Username string `json:"username"`
// 					} `json:"user_created"`
// 					DateCreated time.Time `json:"date_created"`
// 					DateUpdated time.Time `json:"date_updated"`
// 					Files       []struct {
// 						ID       string `json:"id"`
// 						FileType string `json:"file_type"`
// 						Movie    struct {
// 							ID string `json:"id"`
// 						} `json:"movie"`
// 					} `json:"files"`
// 				} `json:"collection_sets"`
// 			} `json:"collection_id"`
// 			MovieSets []struct {
// 				ID          string `json:"id"`
// 				UserCreated struct {
// 					Username string `json:"username"`
// 				} `json:"user_created"`
// 				DateCreated time.Time `json:"date_created"`
// 				DateUpdated time.Time `json:"date_updated"`
// 				Files       []struct {
// 					ID       string `json:"id"`
// 					FileType string `json:"file_type"`
// 					Movie    struct {
// 						ID string `json:"id"`
// 					} `json:"movie"`
// 				} `json:"files"`
// 			} `json:"movie_sets"`
// 		} `json:"movies"`
// 	} `json:"data"`
// }

// type MediuxShowSetsResponse struct {
// 	Data struct {
// 		Shows []struct {
// 			ID           string    `json:"id"`
// 			DateUpdated  time.Time `json:"date_updated"`
// 			Status       string    `json:"status"`
// 			Title        string    `json:"title"`
// 			Tagline      string    `json:"tagline"`
// 			FirstAirDate string    `json:"first_air_date"`
// 			TvdbID       string    `json:"tvdb_id"`
// 			ImdbID       string    `json:"imdb_id"`
// 			TraktID      string    `json:"trakt_id"`
// 			Slug         string    `json:"slug"`
// 			ShowSets     []struct {
// 				ID          string `json:"id"`
// 				UserCreated struct {
// 					Username string `json:"username"`
// 				} `json:"user_created"`
// 				DateCreated time.Time `json:"date_created"`
// 				DateUpdated time.Time `json:"date_updated"`
// 				Files       []struct {
// 					ID       string `json:"id"`
// 					FileType string `json:"file_type"`
// 					Season   *struct {
// 						SeasonNumber int `json:"season_number,omitempty"`
// 					} `json:"season,omitempty"`
// 					Episode *struct {
// 						EpisodeTitle  string `json:"episode_title,omitempty"`
// 						EpisodeNumber int    `json:"episode_number,omitempty"`
// 						SeasonID      *struct {
// 							SeasonNumber int `json:"season_number,omitempty"`
// 						} `json:"season_id,omitempty"`
// 					} `json:"episode,omitempty"`
// 				} `json:"files"`
// 			} `json:"show_sets"`
// 		} `json:"shows"`
// 	} `json:"data"`
// }
