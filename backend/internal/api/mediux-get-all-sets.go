package api

import (
	"aura/internal/logging"
	"context"
	"strconv"
)

func Mediux_FetchAllSets(ctx context.Context, tmdbID, itemType, librarySection string) ([]PosterSet, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetch All Sets from MediUX", logging.LevelInfo)
	defer logAction.Complete()

	actionGenerateRequestBody := logAction.AddSubAction("Generate Request Body", logging.LevelDebug)
	// Generate the request body
	var requestBody map[string]any
	switch itemType {
	case "movie":
		requestBody = Mediux_GenerateMovieRequestBody(tmdbID)
	case "show":
		requestBody = Mediux_GenerateShowRequestBody(tmdbID)
	default:
		actionGenerateRequestBody.SetError("Invalid Item Type", "The item type provided is not valid",
			map[string]any{
				"itemType": itemType,
			})
		return nil, logging.LogErrorInfo{}
	}
	actionGenerateRequestBody.Complete()

	// Send the GraphQL request
	resp, Err := Mediux_SendGraphQLRequest(ctx, requestBody)
	if Err.Message != "" {
		Err.Detail["tmdbID"] = tmdbID
		Err.Detail["itemType"] = itemType
		Err.Detail["librarySection"] = librarySection
		return nil, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxResponse
	Err = DecodeJSONBody(ctx, resp.Body(), &responseBody, "MediuxResponse")
	if Err.Message != "" {
		return nil, Err
	}

	if len(responseBody.Errors) > 0 {
		logAction.SetError("Errors returned from MediUX API", "Review the errors for more details",
			map[string]any{
				"tmdbID":   tmdbID,
				"itemType": itemType,
				"errors":   responseBody.Errors,
			})
		return nil, *logAction.Error
	}

	// Check Response is Valid
	actionValidateResponse := logAction.AddSubAction("Validate MediUX Response", logging.LevelDebug)
	// Check if the response is nil on all fields
	switch itemType {
	case "movie":
		if responseBody.Data.Movie == nil {
			actionValidateResponse.SetError("Movie not found in the response", "Ensure the TMDB ID is correct and the movie exists in the MediUX database.",
				map[string]any{
					"tmdbID":   tmdbID,
					"itemType": itemType,
				})
			return nil, *actionValidateResponse.Error
		}

		if responseBody.Data.Movie.ID == "" {
			actionValidateResponse.SetError("Movie ID not found in the response", "Ensure the TMDB ID is correct and the movie exists in the MediUX database.",
				map[string]any{
					"tmdbID":   tmdbID,
					"itemType": itemType,
				})
			return nil, *actionValidateResponse.Error
		}

		if responseBody.Data.Movie.CollectionID == nil &&
			responseBody.Data.Movie.Posters == nil &&
			responseBody.Data.Movie.Backdrops == nil {
			actionValidateResponse.SetError("Movie sets not found in the response", "Ensure the TMDB ID is correct and the movie has sets in the MediUX database.",
				map[string]any{
					"tmdbID":   tmdbID,
					"itemType": itemType,
				})
			return nil, *actionValidateResponse.Error
		}

	case "show":
		if responseBody.Data.Show == nil {
			actionValidateResponse.SetError("Show not found in the response", "Ensure the TMDB ID is correct and the show exists in the MediUX database.",
				map[string]any{
					"tmdbID":   tmdbID,
					"itemType": itemType,
				})
			return nil, *actionValidateResponse.Error
		}
		if responseBody.Data.Show.ID == "" {
			actionValidateResponse.SetError("Show ID not found in the response", "Ensure the TMDB ID is correct and the show exists in the MediUX database.",
				map[string]any{
					"tmdbID":   tmdbID,
					"itemType": itemType,
				})
			return nil, *actionValidateResponse.Error
		}

		if responseBody.Data.Show.Posters == nil &&
			responseBody.Data.Show.Backdrops == nil &&
			responseBody.Data.Show.Seasons == nil {
			actionValidateResponse.SetError("Show sets not found in the response", "Ensure the TMDB ID is correct and the show has sets in the MediUX database.",
				map[string]any{
					"tmdbID":   tmdbID,
					"itemType": itemType,
				})
			return nil, *actionValidateResponse.Error
		}
	}

	var posterSets []PosterSet
	var itemTitle string
	switch itemType {
	case "movie":
		posterSets = processMovieResponse(ctx, librarySection, tmdbID, (*responseBody.Data.Movie))
		itemTitle = responseBody.Data.Movie.Title
	case "show":
		posterSets = processShowResponse(ctx, librarySection, tmdbID, (*responseBody.Data.Show))
		itemTitle = responseBody.Data.Show.Title
	}

	if len(posterSets) == 0 {
		actionValidateResponse.Status = logging.StatusWarn
		actionValidateResponse.SetError("No poster sets found",
			"Use the MediUX site to confirm that poster sets exist for this item.",
			map[string]any{
				"tmdbID":   tmdbID,
				"itemType": itemType,
				"title":    itemTitle,
			})
		return nil, *actionValidateResponse.Error
	}
	actionValidateResponse.Complete()

	return posterSets, logging.LogErrorInfo{}
}

func processShowResponse(ctx context.Context, librarySection, tmdbID string, show MediuxShowByID) []PosterSet {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Process Show Response", logging.LevelTrace)
	defer logAction.Complete()

	showSetMap := make(map[string]*PosterSet)

	cachedItem, exists := Global_Cache_LibraryStore.GetMediaItemFromSectionByTMDBID(librarySection, tmdbID)
	if !exists {
		logAction.SetError("Could not find show in cache", "Try refreshing on the Home page", map[string]any{
			"title":  show.Title,
			"tmdbID": tmdbID,
		})
		return []PosterSet{}
	}

	var baseMinMediaItem = MediaItem{
		TMDB_ID:         tmdbID,
		LibraryTitle:    cachedItem.LibraryTitle,
		RatingKey:       cachedItem.RatingKey,
		Type:            "show",
		Title:           cachedItem.Title,
		Year:            cachedItem.Year,
		ExistInDatabase: cachedItem.ExistInDatabase,
		DBSavedSets:     cachedItem.DBSavedSets,
	}

	// If the show struct is basically empty, short-circuit
	if show.ID == "" &&
		len(show.Posters) == 0 &&
		len(show.Backdrops) == 0 &&
		len(show.Seasons) == 0 {
		logAction.SetError("Show response contained no assets", "Ensure the TMDB ID is correct and the show has assets in the MediUX database.",
			map[string]any{
				"title":    show.Title,
				"tmdbID":   tmdbID,
				"itemType": "show",
			})
		return []PosterSet{}
	}

	if show.ID != tmdbID {
		logAction.AppendWarning("message", "TMDB ID mismatch between request and response")
		logAction.AppendResult("requested", tmdbID)
		logAction.AppendResult("got", show.ID)
		return []PosterSet{}
	}

	var showResultDetails = map[string]any{
		"id":             show.ID,
		"title":          show.Title,
		"posters":        len(show.Posters),
		"backdrops":      len(show.Backdrops),
		"season_posters": 0,
		"titlecards":     0,
	}
	if len(show.Posters) > 0 {
		for _, poster := range show.Posters {
			if poster.ShowSet != nil && poster.ShowSet.ID != "" {
				setInfo := poster.ShowSet
				newPoster := PosterFile{
					ID:       poster.ID,
					Type:     "poster",
					Modified: poster.ModifiedOn,
					FileSize: parseFileSize(poster.FileSize),
					Src:      poster.Src,
					Blurhash: poster.Blurhash,
					Show: &PosterFileShow{
						ID:        show.ID,
						Title:     show.Title,
						MediaItem: baseMinMediaItem,
					},
				}

				if ps, exists := showSetMap[setInfo.ID]; exists {
					ps.Poster = &newPoster
				} else {
					newPosterSet := &PosterSet{
						ID:                setInfo.ID,
						Title:             setInfo.SetTitle,
						Type:              "show",
						User:              SetUser{Name: setInfo.UserCreated.Username},
						DateCreated:       setInfo.DateCreated,
						DateUpdated:       setInfo.DateUpdated,
						Status:            show.Status,
						TMDB_PosterPath:   show.TMDB_Poster,
						TMDB_BackdropPath: show.TMDB_Backdrop,
					}
					newPosterSet.Poster = &newPoster
					showSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}
	if len(show.Backdrops) > 0 {
		for _, backdrop := range show.Backdrops {
			if backdrop.ShowSet != nil && backdrop.ShowSet.ID != "" {
				setInfo := backdrop.ShowSet
				newBackdrop := PosterFile{
					ID:       backdrop.ID,
					Type:     "backdrop",
					Modified: backdrop.ModifiedOn,
					FileSize: parseFileSize(backdrop.FileSize),
					Src:      backdrop.Src,
					Blurhash: backdrop.Blurhash,
					Show: &PosterFileShow{
						ID:        show.ID,
						Title:     show.Title,
						MediaItem: baseMinMediaItem,
					},
				}
				if ps, exists := showSetMap[setInfo.ID]; exists {
					ps.Backdrop = &newBackdrop
				} else {
					newPosterSet := &PosterSet{
						ID:                setInfo.ID,
						Title:             setInfo.SetTitle,
						Type:              "show",
						User:              SetUser{Name: setInfo.UserCreated.Username},
						DateCreated:       setInfo.DateCreated,
						DateUpdated:       setInfo.DateUpdated,
						Status:            show.Status,
						TMDB_PosterPath:   show.TMDB_Poster,
						TMDB_BackdropPath: show.TMDB_Backdrop,
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
		showResultDetails["season_posters"] = totalSeasonPosters
		showResultDetails["titlecards"] = totalTitlecards
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
						seasonItem = &baseMinMediaItem // Use existing cached item
					}
					newPoster := PosterFile{
						ID:       poster.ID,
						Type:     seasonType,
						Modified: poster.ModifiedOn,
						FileSize: parseFileSize(poster.FileSize),
						Src:      poster.Src,
						Blurhash: poster.Blurhash,
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
							ID:                setInfo.ID,
							Title:             setInfo.SetTitle,
							Type:              "show",
							User:              SetUser{Name: setInfo.UserCreated.Username},
							DateCreated:       setInfo.DateCreated,
							DateUpdated:       setInfo.DateUpdated,
							Status:            show.Status,
							TMDB_PosterPath:   show.TMDB_Poster,
							TMDB_BackdropPath: show.TMDB_Backdrop,
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
							episodeItem = &baseMinMediaItem // Use existing cached item
						}

						newTitlecard := PosterFile{
							ID:       titlecard.ID,
							Type:     "titlecard",
							Modified: titlecard.ModifiedOn,
							FileSize: parseFileSize(titlecard.FileSize),
							Src:      titlecard.Src,
							Blurhash: titlecard.Blurhash,
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
								ID:                setInfo.ID,
								Title:             setInfo.SetTitle,
								Type:              "show",
								User:              SetUser{Name: setInfo.UserCreated.Username},
								DateCreated:       setInfo.DateCreated,
								DateUpdated:       setInfo.DateUpdated,
								Status:            show.Status,
								TMDB_PosterPath:   show.TMDB_Poster,
								TMDB_BackdropPath: show.TMDB_Backdrop,
							}
							newPosterSet.TitleCards = []PosterFile{newTitlecard}
							showSetMap[setInfo.ID] = newPosterSet
						}

					}

				}
			}
		}
	}
	logAction.AppendResult("show", showResultDetails)

	// Convert the map to a slice
	var posterSets []PosterSet
	for _, set := range showSetMap {
		posterSets = append(posterSets, *set)
	}

	return posterSets
}

func processMovieResponse(ctx context.Context, librarySection, tmdbID string, movie MediuxMovieByID) []PosterSet {
	var posterSets []PosterSet

	if movie.CollectionID != nil {
		posterSets = append(posterSets, processMovieCollection(ctx, tmdbID, librarySection, movie.ID, *movie.CollectionID)...)
	}
	posterSets = append(posterSets, processMovieSetPostersAndBackdrops(ctx, librarySection, tmdbID, movie)...)

	return posterSets
}

func processMovieSetPostersAndBackdrops(ctx context.Context, librarySection string, tmdbID string, movie MediuxMovieByID) []PosterSet {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Process Movie Set Posters and Backdrops", logging.LevelTrace)
	defer logAction.Complete()

	var posterSets []PosterSet
	movieSetMap := make(map[string]*PosterSet)

	cachedItem, exists := Global_Cache_LibraryStore.GetMediaItemFromSectionByTMDBID(librarySection, tmdbID)
	if !exists {
		logAction.SetError("Could not find movie in cache", "Try refreshing on the Home page", map[string]any{
			"title":  movie.Title,
			"tmdbID": tmdbID,
		})
		return nil
	}

	baseMinMediaItem := MediaItem{
		TMDB_ID:         tmdbID,
		LibraryTitle:    cachedItem.LibraryTitle,
		RatingKey:       cachedItem.RatingKey,
		Type:            "movie",
		Title:           cachedItem.Title,
		Year:            cachedItem.Year,
		ExistInDatabase: cachedItem.ExistInDatabase,
		DBSavedSets:     cachedItem.DBSavedSets,
	}

	logAction.AppendResult("movie", map[string]any{
		"id":        movie.ID,
		"title":     movie.Title,
		"posters":   len(movie.Posters),
		"backdrops": len(movie.Backdrops),
	})

	if len(movie.Posters) > 0 {
		for _, poster := range movie.Posters {
			if poster.MovieSet != nil && poster.MovieSet.ID != "" {
				setInfo := poster.MovieSet
				newPoster := PosterFile{
					ID:       poster.ID,
					Type:     "poster",
					Modified: poster.ModifiedOn,
					FileSize: parseFileSize(poster.FileSize),
					Src:      poster.Src,
					Blurhash: poster.Blurhash,
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
						MediaItem:   baseMinMediaItem,
					},
				}

				// Check to see this set already exists in the map
				if ps, exists := movieSetMap[setInfo.ID]; exists {
					ps.Poster = &newPoster
				} else {
					newPosterSet := &PosterSet{
						ID:                setInfo.ID,
						Title:             setInfo.SetTitle,
						Type:              "movie",
						User:              SetUser{Name: setInfo.UserCreated.Username},
						DateCreated:       setInfo.DateCreated,
						DateUpdated:       setInfo.DateUpdated,
						Status:            movie.Status,
						TMDB_PosterPath:   movie.TMDB_Poster,
						TMDB_BackdropPath: movie.TMDB_Backdrop,
					}
					newPosterSet.Poster = &newPoster
					movieSetMap[setInfo.ID] = newPosterSet
				}
			}
		}
	}

	if len(movie.Backdrops) > 0 {
		for _, backdrop := range movie.Backdrops {
			if backdrop.MovieSet != nil && backdrop.MovieSet.ID != "" {
				setInfo := backdrop.MovieSet
				newBackdrop := PosterFile{
					ID:       backdrop.ID,
					Type:     "backdrop",
					Modified: backdrop.ModifiedOn,
					FileSize: parseFileSize(backdrop.FileSize),
					Src:      backdrop.Src,
					Blurhash: backdrop.Blurhash,
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
						MediaItem:   baseMinMediaItem,
					},
				}
				if ps, exists := movieSetMap[setInfo.ID]; exists {
					ps.Backdrop = &newBackdrop
				} else {
					newPosterSet := &PosterSet{
						ID:                setInfo.ID,
						Title:             setInfo.SetTitle,
						Type:              "movie",
						User:              SetUser{Name: setInfo.UserCreated.Username},
						DateCreated:       setInfo.DateCreated,
						DateUpdated:       setInfo.DateUpdated,
						Status:            movie.Status,
						TMDB_PosterPath:   movie.TMDB_Poster,
						TMDB_BackdropPath: movie.TMDB_Backdrop,
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

func processMovieCollection(ctx context.Context, tmdbID, librarySection, mainMovieID string, collection MediuxMovieCollectionID) []PosterSet {
	movies := collection.Movies
	if len(movies) == 0 {
		return nil
	}

	ctx, logAction := logging.AddSubActionToContext(ctx, "Process Movie Collection", logging.LevelTrace)
	defer logAction.Complete()

	logAction.AppendResult("number_of_movies", len(movies))
	collectionSetMap := make(map[string]*PosterSet)
	var movieResults []map[string]any
	for _, movie := range movies {
		movieResult := map[string]any{
			"id":        movie.ID,
			"title":     movie.Title,
			"posters":   len(movie.Posters),
			"backdrops": len(movie.Backdrops),
		}
		movieResults = append(movieResults, movieResult)

		cachedItem, _ := Global_Cache_LibraryStore.GetMediaItemFromSectionByTMDBID(librarySection, movie.ID)
		if cachedItem == nil {
			logAction.SetError("Could not find movie in cache", "Try refreshing on the Home page", map[string]any{
				"title":  movie.Title,
				"tmdbID": tmdbID,
			})
		}

		baseMinMediaItem := MediaItem{
			TMDB_ID:         movie.ID,
			LibraryTitle:    cachedItem.LibraryTitle,
			RatingKey:       cachedItem.RatingKey,
			Type:            "movie",
			Title:           cachedItem.Title,
			Year:            cachedItem.Year,
			ExistInDatabase: cachedItem.ExistInDatabase,
			DBSavedSets:     cachedItem.DBSavedSets,
		}

		if len(movie.Posters) > 0 {
			for _, poster := range movie.Posters {
				if poster.CollectionSet.ID != "" {
					setInfo := poster.CollectionSet

					newPoster := PosterFile{
						ID:       poster.ID,
						Type:     "poster",
						Modified: poster.ModifiedOn,
						FileSize: parseFileSize(poster.FileSize),
						Src:      poster.Src,
						Blurhash: poster.Blurhash,
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
							MediaItem:   baseMinMediaItem,
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
							ID:                setInfo.ID,
							Title:             setInfo.SetTitle,
							Type:              "collection",
							User:              SetUser{Name: setInfo.UserCreated.Username},
							DateCreated:       setInfo.DateCreated,
							DateUpdated:       setInfo.DateUpdated,
							Status:            movie.Status,
							TMDB_PosterPath:   movie.TMDB_Poster,
							TMDB_BackdropPath: movie.TMDB_Backdrop,
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
			for _, backdrop := range movie.Backdrops {
				if backdrop.CollectionSet.ID != "" {
					setInfo := backdrop.CollectionSet

					newBackdrop := PosterFile{
						ID:       backdrop.ID,
						Type:     "backdrop",
						Modified: backdrop.ModifiedOn,
						FileSize: parseFileSize(backdrop.FileSize),
						Src:      backdrop.Src,
						Blurhash: backdrop.Blurhash,
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
							MediaItem:   baseMinMediaItem,
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
							ID:                setInfo.ID,
							Title:             setInfo.SetTitle,
							Type:              "collection",
							User:              SetUser{Name: setInfo.UserCreated.Username},
							DateCreated:       setInfo.DateCreated,
							DateUpdated:       setInfo.DateUpdated,
							Status:            movie.Status,
							TMDB_PosterPath:   movie.TMDB_Poster,
							TMDB_BackdropPath: movie.TMDB_Backdrop,
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
	logAction.AppendResult("movies", movieResults)

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
