package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
)

func Mediux_FetchAllSets(tmdbID, itemType, librarySection string) ([]PosterSet, logging.StandardError) {
	Err := logging.NewStandardError()

	// Generate the request body
	var requestBody map[string]any
	switch itemType {
	case "movie":
		requestBody = Mediux_GenerateMovieRequestBody(tmdbID)
	case "show":
		requestBody = Mediux_GenerateShowRequestBody(tmdbID)
	default:
		Err.Message = "Invalid item type provided"
		Err.HelpText = "Ensure the item type is either 'movie' or 'show'."
		Err.Details = map[string]any{
			"itemType": itemType,
		}
		return nil, Err
	}
	// Create a new Resty client
	client := resty.New()

	// Send the GraphQL request to the Mediux API as a POST request
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", Global_Config.Mediux.Token)).
		SetBody(requestBody).
		Post("https://images.mediux.io/graphql")
	if err != nil {
		Err.Message = "Failed to make request to Mediux API"
		Err.HelpText = "Ensure the Mediux API is reachable and the token is valid."
		Err.Details = map[string]any{
			"error":          err.Error(),
			"tmdbID":         tmdbID,
			"itemType":       itemType,
			"librarySection": librarySection,
			"responseBody":   string(response.Body()),
			"statusCode":     response.StatusCode(),
		}
		return nil, Err
	}
	if response.StatusCode() != http.StatusOK {
		Err.Message = "Mediux API returned non-OK status"
		Err.HelpText = "Check the Mediux API status or your request parameters."
		Err.Details = map[string]any{
			"tmdbID":         tmdbID,
			"itemType":       itemType,
			"librarySection": librarySection,
			"responseBody":   string(response.Body()),
			"statusCode":     response.StatusCode(),
		}
		return nil, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxResponse

	err = json.Unmarshal(response.Body(), &responseBody)
	if err != nil {
		Err.Message = "Failed to parse Mediux API response"
		Err.HelpText = "Ensure the Mediux API is returning a valid JSON response."
		Err.Details = map[string]any{
			"error":          err.Error(),
			"tmdbID":         tmdbID,
			"itemType":       itemType,
			"librarySection": librarySection,
			"responseBody":   string(response.Body()),
			"statusCode":     response.StatusCode(),
		}
		return nil, Err
	}

	// Check if the response is nil on all fields
	switch itemType {
	case "movie":
		if responseBody.Data.Movie.ID == "" {
			Err.Message = "Movie not found in the response"
			Err.HelpText = "Ensure the TMDB ID is correct and the movie exists in the Mediux database."
			Err.Details = map[string]any{
				"tmdbID":   tmdbID,
				"itemType": itemType,
			}
			return nil, Err
		}

		if responseBody.Data.Movie.CollectionID == nil &&
			responseBody.Data.Movie.Posters == nil &&
			responseBody.Data.Movie.Backdrops == nil {
			Err.Message = "Movie sets not found in the response"
			Err.HelpText = "Ensure the TMDB ID is correct and the movie has sets in the Mediux database."
			Err.Details = map[string]any{
				"tmdbID":   tmdbID,
				"itemType": itemType,
			}
			return nil, Err
		}

	case "show":
		if responseBody.Data.Show.ID == "" {
			Err.Message = "Show not found in the response"
			Err.HelpText = "Ensure the TMDB ID is correct and the show exists in the Mediux database."
			Err.Details = map[string]any{
				"tmdbID":   tmdbID,
				"itemType": itemType,
			}
			return nil, Err
		}

		if responseBody.Data.Show.Posters == nil &&
			responseBody.Data.Show.Backdrops == nil &&
			responseBody.Data.Show.Seasons == nil {
			Err.Message = "Show sets not found in the response"
			Err.HelpText = "Ensure the TMDB ID is correct and the show has sets in the Mediux database."
			Err.Details = map[string]any{
				"tmdbID":   tmdbID,
				"itemType": itemType,
			}
			return nil, Err
		}
	}

	var posterSets []PosterSet
	switch itemType {
	case "movie":
		posterSets = processMovieResponse(librarySection, tmdbID, (*responseBody.Data.Movie))
	case "show":
		posterSets = processShowResponse(librarySection, tmdbID, (*responseBody.Data.Show))
	}

	return posterSets, logging.StandardError{}
}

func processShowResponse(librarySection, tmdbID string, show MediuxShowByID) []PosterSet {
	logging.LOG.Trace(fmt.Sprintf("Processing Show Set for - %s (%s)", show.Title, tmdbID))
	showSetMap := make(map[string]*PosterSet)

	cachedItem, exists := Global_Cache_LibraryStore.GetMediaItemFromSection(librarySection, tmdbID)
	if !exists {
		logging.LOG.Error(fmt.Sprintf("\tCould not find '%s' in cache", show.Title))
		return []PosterSet{}
	}

	// If the show struct is basically empty, short-circuit
	if show.ID == "" &&
		len(show.Posters) == 0 &&
		len(show.Backdrops) == 0 &&
		len(show.Seasons) == 0 {
		logging.LOG.Trace("\tShow response contained no assets")
		return []PosterSet{}
	}

	if show.ID != tmdbID {
		logging.LOG.Warn(fmt.Sprintf("\tTMDB ID mismatch: requested %s but got %s", tmdbID, show.ID))
		return []PosterSet{}
	}

	if len(show.Posters) > 0 {
		logging.LOG.Trace(fmt.Sprintf("\tFound %d posters", len(show.Posters)))
		for _, poster := range show.Posters {
			if poster.ShowSet != nil && poster.ShowSet.ID != "" {
				setInfo := poster.ShowSet
				newPoster := PosterFile{
					ID:       poster.ID,
					Type:     "poster",
					Modified: poster.ModifiedOn,
					FileSize: parseFileSize(poster.FileSize),
					Show: &PosterFileShow{
						ID:        show.ID,
						Title:     show.Title,
						MediaItem: *cachedItem,
					},
				}

				if ps, exists := showSetMap[setInfo.ID]; exists {
					ps.Poster = &newPoster
				} else {
					newPosterSet := &PosterSet{
						ID:          setInfo.ID,
						Title:       setInfo.SetTitle,
						Type:        "show",
						User:        SetUser{Name: setInfo.UserCreated.Username},
						DateCreated: setInfo.DateCreated,
						DateUpdated: setInfo.DateUpdated,
						Status:      show.Status,
					}
					newPosterSet.Poster = &newPoster
					showSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}
	if len(show.Backdrops) > 0 {
		logging.LOG.Trace(fmt.Sprintf("\tFound %d backdrops", len(show.Backdrops)))
		for _, backdrop := range show.Backdrops {
			if backdrop.ShowSet != nil && backdrop.ShowSet.ID != "" {
				setInfo := backdrop.ShowSet
				newBackdrop := PosterFile{
					ID:       backdrop.ID,
					Type:     "backdrop",
					Modified: backdrop.ModifiedOn,
					FileSize: parseFileSize(backdrop.FileSize),
					Show: &PosterFileShow{
						ID:        show.ID,
						Title:     show.Title,
						MediaItem: *cachedItem,
					},
				}
				if ps, exists := showSetMap[setInfo.ID]; exists {
					ps.Backdrop = &newBackdrop
				} else {
					newPosterSet := &PosterSet{
						ID:          setInfo.ID,
						Title:       setInfo.SetTitle,
						Type:        "show",
						User:        SetUser{Name: setInfo.UserCreated.Username},
						DateCreated: setInfo.DateCreated,
						DateUpdated: setInfo.DateUpdated,
						Status:      show.Status,
					}
					newPosterSet.Backdrop = &newBackdrop
					showSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}
	if len(show.Seasons) > 0 {
		var totalSeasonPosters, totalTitlecards int
		for _, season := range show.Seasons {
			totalSeasonPosters += len(season.Posters)
			for _, episode := range season.Episodes {
				totalTitlecards += len(episode.Titlecards)
			}
		}
		logging.LOG.Trace(fmt.Sprintf("\tFound %d season posters", totalSeasonPosters))
		logging.LOG.Trace(fmt.Sprintf("\tFound %d titlecards", totalTitlecards))
		for _, season := range show.Seasons {
			for _, poster := range season.Posters {
				if poster.ShowSet != nil && poster.ShowSet.ID != "" {
					setInfo := poster.ShowSet
					var seasonType string
					if season.SeasonNumber == 0 {
						seasonType = "specialSeasonPoster"
					} else {
						seasonType = "seasonPoster"
					}
					// Check if the Season exists in the cachedItem
					seasonExists := false
					for _, cachedSeason := range cachedItem.Series.Seasons {
						if cachedSeason.SeasonNumber == season.SeasonNumber {
							seasonExists = true
							break
						}
					}
					var seasonItem *MediaItem
					if !seasonExists {
						seasonItem = &MediaItem{} // Create a new empty MediaItem
					} else {
						seasonItem = cachedItem // Use existing cached item
					}
					newPoster := PosterFile{
						ID:       poster.ID,
						Type:     seasonType,
						Modified: poster.ModifiedOn,
						FileSize: parseFileSize(poster.FileSize),
						Show: &PosterFileShow{
							ID:        show.ID,
							Title:     show.Title,
							MediaItem: *seasonItem,
						},
						Season: &PosterFileSeason{
							Number: season.SeasonNumber,
						},
					}
					if ps, exists := showSetMap[setInfo.ID]; exists {
						ps.SeasonPosters = append(ps.SeasonPosters, newPoster)
					} else {
						newPosterSet := &PosterSet{
							ID:          setInfo.ID,
							Title:       setInfo.SetTitle,
							Type:        "show",
							User:        SetUser{Name: setInfo.UserCreated.Username},
							DateCreated: setInfo.DateCreated,
							DateUpdated: setInfo.DateUpdated,
							Status:      show.Status,
						}
						newPosterSet.SeasonPosters = []PosterFile{newPoster}
						showSetMap[setInfo.ID] = newPosterSet
					}
				}
			}
			for _, episode := range season.Episodes {
				for _, titlecard := range episode.Titlecards {
					if titlecard.ShowSet != nil && titlecard.ShowSet.ID != "" {
						setInfo := titlecard.ShowSet

						// Check if the Episode exists in the cachedItem
						episodeExists := false
						for _, cachedSeason := range cachedItem.Series.Seasons {
							for _, cachedEpisode := range cachedSeason.Episodes {
								if cachedEpisode.SeasonNumber == episode.Season.SeasonNumber &&
									cachedEpisode.EpisodeNumber == episode.EpisodeNumber {
									episodeExists = true
									break
								}
							}
						}
						var episodeItem *MediaItem
						if !episodeExists {
							episodeItem = &MediaItem{} // Create a new empty MediaItem
						} else {
							episodeItem = cachedItem // Use existing cached item
						}

						newTitlecard := PosterFile{
							ID:       titlecard.ID,
							Type:     "titlecard",
							Modified: titlecard.ModifiedOn,
							FileSize: parseFileSize(titlecard.FileSize),
							Show: &PosterFileShow{
								ID:        show.ID,
								Title:     show.Title,
								MediaItem: *episodeItem,
							},
							Episode: &PosterFileEpisode{
								Title:         episode.EpisodeTitle,
								EpisodeNumber: episode.EpisodeNumber,
								SeasonNumber:  episode.Season.SeasonNumber,
							},
						}
						if ps, exists := showSetMap[setInfo.ID]; exists {
							ps.TitleCards = append(ps.TitleCards, newTitlecard)
						} else {
							newPosterSet := &PosterSet{
								ID:          setInfo.ID,
								Title:       setInfo.SetTitle,
								Type:        "show",
								User:        SetUser{Name: setInfo.UserCreated.Username},
								DateCreated: setInfo.DateCreated,
								DateUpdated: setInfo.DateUpdated,
								Status:      show.Status,
							}
							newPosterSet.TitleCards = []PosterFile{newTitlecard}
							showSetMap[setInfo.ID] = newPosterSet
						}

					}

				}
			}
		}
	}

	// Convert the map to a slice
	var posterSets []PosterSet
	for _, set := range showSetMap {
		posterSets = append(posterSets, *set)
	}

	return posterSets

}

func processMovieResponse(librarySection, tmdbID string, movie MediuxMovieByID) []PosterSet {
	var posterSets []PosterSet

	if movie.CollectionID != nil {
		posterSets = append(posterSets, processMovieCollection(tmdbID, librarySection, movie.ID, *movie.CollectionID)...)
	}
	posterSets = append(posterSets, processMovieSetPostersAndBackdrops(librarySection, tmdbID, movie)...)

	return posterSets
}

func processMovieSetPostersAndBackdrops(librarySection string, tmdbID string, movie MediuxMovieByID) []PosterSet {
	logging.LOG.Trace(fmt.Sprintf("Processing Movie Set for - %s (%s)", movie.Title, tmdbID))
	var posterSets []PosterSet
	movieSetMap := make(map[string]*PosterSet)

	cachedItem, exists := Global_Cache_LibraryStore.GetMediaItemFromSection(librarySection, tmdbID)
	if !exists {
		logging.LOG.Error(fmt.Sprintf("\tCould not find '%s' in cache", movie.Title))
		return nil
	}

	if len(movie.Posters) > 0 {
		logging.LOG.Trace(fmt.Sprintf("\tFound %d posters", len(movie.Posters)))
		for _, poster := range movie.Posters {
			if poster.MovieSet != nil && poster.MovieSet.ID != "" {
				setInfo := poster.MovieSet
				newPoster := PosterFile{
					ID:       poster.ID,
					Type:     "poster",
					Modified: poster.ModifiedOn,
					FileSize: parseFileSize(poster.FileSize),
					Movie: &PosterFileMovie{
						ID:          movie.ID,
						Title:       movie.Title,
						Status:      movie.Status,
						Tagline:     movie.Tagline,
						Slug:        movie.Slug,
						DateUpdated: movie.DateUpdated,
						TvdbID:      movie.TvdbID,
						ImdbID:      movie.ImdbID,
						TraktID:     movie.TraktID,
						ReleaseDate: movie.ReleaseDate,
						MediaItem:   *cachedItem,
					},
				}

				// Check to see this set already exists in the map
				if ps, exists := movieSetMap[setInfo.ID]; exists {
					ps.Poster = &newPoster
				} else {
					newPosterSet := &PosterSet{
						ID:          setInfo.ID,
						Title:       setInfo.SetTitle,
						Type:        "movie",
						User:        SetUser{Name: setInfo.UserCreated.Username},
						DateCreated: setInfo.DateCreated,
						DateUpdated: setInfo.DateUpdated,
						Status:      movie.Status,
					}
					newPosterSet.Poster = &newPoster
					movieSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}

	if len(movie.Backdrops) > 0 {
		logging.LOG.Trace(fmt.Sprintf("\tFound %d backdrops", len(movie.Backdrops)))
		for _, backdrop := range movie.Backdrops {
			if backdrop.MovieSet != nil && backdrop.MovieSet.ID != "" {
				setInfo := backdrop.MovieSet
				newBackdrop := PosterFile{
					ID:       backdrop.ID,
					Type:     "backdrop",
					Modified: backdrop.ModifiedOn,
					FileSize: parseFileSize(backdrop.FileSize),
					Movie: &PosterFileMovie{
						ID:          movie.ID,
						Title:       movie.Title,
						Status:      movie.Status,
						Tagline:     movie.Tagline,
						Slug:        movie.Slug,
						DateUpdated: movie.DateUpdated,
						TvdbID:      movie.TvdbID,
						ImdbID:      movie.ImdbID,
						TraktID:     movie.TraktID,
						ReleaseDate: movie.ReleaseDate,
						MediaItem:   *cachedItem,
					},
				}
				if ps, exists := movieSetMap[setInfo.ID]; exists {
					ps.Backdrop = &newBackdrop
				} else {
					newPosterSet := &PosterSet{
						ID:          setInfo.ID,
						Title:       setInfo.SetTitle,
						Type:        "movie",
						User:        SetUser{Name: setInfo.UserCreated.Username},
						DateCreated: setInfo.DateCreated,
						DateUpdated: setInfo.DateUpdated,
						Status:      movie.Status,
					}
					newPosterSet.Backdrop = &newBackdrop
					movieSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}

	// Convert the map to a slice
	for _, set := range movieSetMap {
		posterSets = append(posterSets, *set)
	}

	return posterSets
}

func processMovieCollection(tmdbID, librarySection, mainMovieID string, collection MediuxMovieCollectionID) []PosterSet {
	movies := collection.Movies
	if len(movies) == 0 {
		return nil
	}
	logging.LOG.Trace(fmt.Sprintf("Processing %d movies in '%s' collection", len(movies), collection.CollectionName))
	collectionSetMap := make(map[string]*PosterSet)
	for _, movie := range movies {
		logging.LOG.Trace(fmt.Sprintf("\tProcessing movie: %s", movie.Title))

		cachedItem, _ := Global_Cache_LibraryStore.GetMediaItemFromSectionByTMDBID(librarySection, movie.ID)
		if cachedItem == nil {
			logging.LOG.Error(fmt.Sprintf("\t\tCould not find '%s' in cache", movie.Title))
		}

		if len(movie.Posters) > 0 {
			logging.LOG.Trace(fmt.Sprintf("\t\tFound %d posters", len(movie.Posters)))
			for _, poster := range movie.Posters {
				if poster.CollectionSet.ID != "" {
					setInfo := poster.CollectionSet

					newPoster := PosterFile{
						ID:       poster.ID,
						Type:     "poster",
						Modified: poster.ModifiedOn,
						FileSize: parseFileSize(poster.FileSize),
						Movie: &PosterFileMovie{
							ID:          movie.ID,
							Title:       movie.Title,
							Status:      movie.Status,
							Tagline:     movie.Tagline,
							Slug:        movie.Slug,
							DateUpdated: movie.DateUpdated,
							TvdbID:      movie.TvdbID,
							ImdbID:      movie.ImdbID,
							TraktID:     movie.TraktID,
							ReleaseDate: movie.ReleaseDate,
							MediaItem:   *cachedItem,
						},
					}

					// Check to see this set already exists in the map
					if cs, exists := collectionSetMap[setInfo.ID]; exists {
						if mainMovieID == movie.ID {
							cs.Poster = &newPoster
						} else {
							cs.OtherPosters = append(cs.OtherPosters, newPoster)
						}
					} else {
						// Create a new PosterSet
						newPosterSet := &PosterSet{
							ID:          setInfo.ID,
							Title:       setInfo.SetTitle,
							Type:        "collection",
							User:        SetUser{Name: setInfo.UserCreated.Username},
							DateCreated: setInfo.DateCreated,
							DateUpdated: setInfo.DateUpdated,
							Status:      movie.Status,
						}
						if mainMovieID == movie.ID {
							newPosterSet.Poster = &newPoster
						} else {
							newPosterSet.OtherPosters = append(newPosterSet.OtherPosters, newPoster)
						}
						collectionSetMap[setInfo.ID] = newPosterSet
					}

				}
			}
		}
		if len(movie.Backdrops) > 0 {
			logging.LOG.Trace(fmt.Sprintf("\t\tFound %d backdrops", len(movie.Backdrops)))
			for _, backdrop := range movie.Backdrops {
				if backdrop.CollectionSet.ID != "" {
					setInfo := backdrop.CollectionSet

					newBackdrop := PosterFile{
						ID:       backdrop.ID,
						Type:     "backdrop",
						Modified: backdrop.ModifiedOn,
						FileSize: parseFileSize(backdrop.FileSize),
						Movie: &PosterFileMovie{
							ID:          movie.ID,
							Title:       movie.Title,
							Status:      movie.Status,
							Tagline:     movie.Tagline,
							Slug:        movie.Slug,
							DateUpdated: movie.DateUpdated,
							TvdbID:      movie.TvdbID,
							ImdbID:      movie.ImdbID,
							TraktID:     movie.TraktID,
							ReleaseDate: movie.ReleaseDate,
							MediaItem:   *cachedItem,
						},
					}

					if cs, exists := collectionSetMap[setInfo.ID]; exists {
						if mainMovieID == movie.ID {
							cs.Backdrop = &newBackdrop
						} else {
							cs.OtherBackdrops = append(cs.OtherBackdrops, newBackdrop)
						}
					} else {
						newPosterSet := &PosterSet{
							ID:          setInfo.ID,
							Title:       setInfo.SetTitle,
							Type:        "collection",
							User:        SetUser{Name: setInfo.UserCreated.Username},
							DateCreated: setInfo.DateCreated,
							DateUpdated: setInfo.DateUpdated,
							Status:      movie.Status,
						}
						if mainMovieID == movie.ID {
							newPosterSet.Backdrop = &newBackdrop
						} else {
							newPosterSet.OtherBackdrops = append(newPosterSet.OtherBackdrops, newBackdrop)
						}
						collectionSetMap[setInfo.ID] = newPosterSet
					}
				}
			}
		}
	}

	// Convert the map to a slice
	var posterSets []PosterSet
	for _, set := range collectionSetMap {
		posterSets = append(posterSets, *set)

	}

	return posterSets
}

func parseFileSize(fileSize string) int64 {
	size, err := strconv.Atoi(fileSize)
	if err != nil {
		return 0
	}
	return int64(size)
}
