package api

import (
	"aura/internal/logging"
	"context"
)

func Mediux_FetchAllCollectionImagesByMovieIDs(ctx context.Context, movieTMDBIDs []string) ([]CollectionSet, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetch All Collection Images By Movie TMDB IDs", logging.LevelInfo)
	defer logAction.Complete()

	logAction.AppendResult("movieTMDBIDs_count", len(movieTMDBIDs))
	logAction.AppendResult("ids", movieTMDBIDs)

	actionGenerateRequestBody := logAction.AddSubAction("Generate Request Body", logging.LevelDebug)
	requestBody := Mediux_GenerateCollectionImagesByMovieIDsBody(movieTMDBIDs)
	actionGenerateRequestBody.Complete()

	// Send the GraphQL request
	resp, Err := Mediux_SendGraphQLRequest(ctx, requestBody)
	if Err.Message != "" {
		Err.Detail["movieTMDBIDs"] = movieTMDBIDs
		return nil, Err
	}

	// Parse the response body into the appropriate struct
	var responseBody MediuxCollectionImagesByMovieIDs_Response
	Err = DecodeJSONBody(ctx, resp.Body(), &responseBody, "MediuxCollectionImagesByMovieIDs_Response")
	if Err.Message != "" {
		return nil, Err
	}

	var collectionSets []CollectionSet
	collectionSetMap := make(map[string]*CollectionSet)

	for _, movie := range responseBody.Data.Movies {
		if movie.CollectionID == nil {
			continue
		}
		if len(movie.CollectionID.Posters) == 0 && len(movie.CollectionID.Backdrops) == 0 {
			continue
		}

		// Now we go through each of the posters
		for _, poster := range movie.CollectionID.Posters {
			// Check if we already have a set for this collection ID
			currentSet, exists := collectionSetMap[poster.CollectionSet.ID]
			if !exists {
				newSet := &CollectionSet{
					ID:        poster.CollectionSet.ID,
					Title:     poster.CollectionSet.SetTitle,
					User:      SetUser{Name: poster.UploadedBy.Username},
					Posters:   []PosterFile{},
					Backdrops: []PosterFile{},
				}
				collectionSetMap[poster.CollectionSet.ID] = newSet
				currentSet = newSet
			}

			// Check if this poster is already in the list of posters
			posterFile := PosterFile{
				ID:       poster.ID,
				Type:     "poster",
				Modified: poster.ModifiedOn,
				FileSize: parseFileSize(poster.FileSize),
				Src:      poster.Src,
				Blurhash: poster.Blurhash,
			}

			if !posterExists(currentSet.Posters, posterFile.ID) {
				currentSet.Posters = append(currentSet.Posters, posterFile)
			}
		}

		// Now we go through each of the backdrops
		for _, backdrop := range movie.CollectionID.Backdrops {
			// Check if we already have a set for this collection ID
			currentSet, exists := collectionSetMap[backdrop.CollectionSet.ID]
			if !exists {
				newSet := &CollectionSet{
					ID:        backdrop.CollectionSet.ID,
					Title:     backdrop.CollectionSet.SetTitle,
					User:      SetUser{Name: backdrop.UploadedBy.Username},
					Posters:   []PosterFile{},
					Backdrops: []PosterFile{},
				}
				collectionSetMap[backdrop.CollectionSet.ID] = newSet
				currentSet = newSet
			}

			// Check if this backdrop is already in the list of backdrops
			backdropFile := PosterFile{
				ID:       backdrop.ID,
				Type:     "backdrop",
				Modified: backdrop.ModifiedOn,
				FileSize: parseFileSize(backdrop.FileSize),
				Src:      backdrop.Src,
				Blurhash: backdrop.Blurhash,
			}

			if !posterExists(currentSet.Backdrops, backdropFile.ID) {
				currentSet.Backdrops = append(currentSet.Backdrops, backdropFile)
			}
		}
	}

	collectionSets = make([]CollectionSet, 0, len(collectionSetMap))
	for _, set := range collectionSetMap {
		collectionSets = append(collectionSets, *set)
	}

	if len(collectionSets) == 0 {
		logAction.AppendResult("collectionSets", "no collection sets found")
		logAction.SetError(
			"No Collection Sets Found",
			"Use the MediUX site to confirm that collection images exist for the provided movie TMDB IDs",
			map[string]any{
				"movieTMDBIDs": movieTMDBIDs,
			},
		)
		return collectionSets, *logAction.Error
	}

	return collectionSets, logging.LogErrorInfo{}
}

func posterExists(posters []PosterFile, id string) bool {
	for _, p := range posters {
		if p.ID == id {
			return true
		}
	}
	return false
}
