package api

import (
	"time"
)

type CollectionSet struct {
	ID        string       `json:"ID"`
	Title     string       `json:"Title"`
	User      SetUser      `json:"User"`
	Posters   []PosterFile `json:"Posters"`
	Backdrops []PosterFile `json:"Backdrops"`
}

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
	Src      string             `json:"Src"`
	Blurhash string             `json:"Blurhash"`
	Movie    *PosterFileMovie   `json:"Movie,omitempty"`
	Show     *PosterFileShow    `json:"Show,omitempty"`
	Season   *PosterFileSeason  `json:"Season,omitempty"`
	Episode  *PosterFileEpisode `json:"Episode,omitempty"`
}

type PosterFileShow struct {
	ID        string    `json:"ID"`
	Title     string    `json:"Title"`
	MediaItem MediaItem `json:"MediaItem"`
}
type PosterFileMovie struct {
	ID          string    `json:"ID"`
	Title       string    `json:"Title"`
	Status      string    `json:"Status,omitempty"`
	Tagline     string    `json:"Tagline,omitempty"`
	Slug        string    `json:"Slug,omitempty"`
	DateUpdated time.Time `json:"DateUpdated"`
	TvdbID      string    `json:"TvdbID,omitempty"`
	ImdbID      string    `json:"ImdbID,omitempty"`
	TraktID     string    `json:"TraktID,omitempty"`
	ReleaseDate string    `json:"ReleaseDate,omitempty"`
	MediaItem   MediaItem `json:"MediaItem"`
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
