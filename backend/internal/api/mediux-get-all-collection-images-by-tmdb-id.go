package api

import (
	"aura/internal/logging"
	"context"
)

func Mediux_FetchAllCollectionImagesByTMDBID(ctx context.Context, tmdbID string) ([]CollectionSet, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetch All Collection Images By TMDB ID", logging.LevelInfo)
	defer logAction.Complete()

	actionGenerateRequestBody := logAction.AddSubAction("Generate Request Body", logging.LevelDebug)
	requestBody := Mediux_GenerateCollectionImagesByTMDBID(tmdbID)
	actionGenerateRequestBody.Complete()

	// Send the GraphQL request
	resp, Err := Mediux_SendGraphQLRequest(ctx, requestBody)
	if Err.Message != "" {
		Err.Detail["tmdbID"] = tmdbID
		return nil, Err
	}

	// Parse the response body into the appropriate struct
	var responseBody MediuxCollectionImageByTMDB_ID_Response
	Err = DecodeJSONBody(ctx, resp.Body(), &responseBody, "MediuxCollectionImageByTMDB_ID_Response")
	if Err.Message != "" {
		return nil, Err
	}

	// If the collection is nil, return empty
	if responseBody.Data.Collection.ID == "" {
		return []CollectionSet{}, logging.LogErrorInfo{}
	}

	var collectionSets []CollectionSet
	collectionSetMap := make(map[string]*CollectionSet)

	// Go through each of the posters and create collection sets
	for _, poster := range responseBody.Data.Collection.Posters {
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

	// Now do the backdrops
	for _, backdrop := range responseBody.Data.Collection.Backdrops {
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
				"tmdbID": tmdbID,
			},
		)
		return collectionSets, *logAction.Error
	}

	return collectionSets, logging.LogErrorInfo{}
}
