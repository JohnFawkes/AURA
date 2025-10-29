package api

import (
	"aura/internal/logging"
	"context"
)

func Mediux_FetchShowSetByID(ctx context.Context, librarySection, tmdbID, setID string) (PosterSet, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetching Show Set by ID from Mediux", logging.LevelTrace)
	defer logAction.Complete()

	// If cache is empty, return false
	if Global_Cache_LibraryStore.IsEmpty() {
		logAction.SetError("Cache is empty", "Cannot fetch show set by ID because the library cache is empty", nil)
		return PosterSet{}, *logAction.Error
	}

	requestBody := Mediux_GenerateShowSetByIDRequestBody(setID)

	// Send the GraphQL request
	resp, Err := Mediux_SendGraphQLRequest(ctx, requestBody)
	if Err.Message != "" {
		Err.Detail["librarySection"] = librarySection
		Err.Detail["tmdbID"] = tmdbID
		Err.Detail["setID"] = setID
		return PosterSet{}, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxShowSetResponse
	Err = DecodeJSONBody(ctx, resp.Body(), &responseBody, "MediuxShowSetResponse")
	if Err.Message != "" {
		return PosterSet{}, Err
	}

	// Validate Show Set
	actionShowSetValidation := logAction.AddSubAction("Validating Show Set Data", logging.LevelTrace)
	showSet := responseBody.Data.ShowSetID
	if showSet.ID == "" {
		actionShowSetValidation.SetError("Invalid Show Set Data", "Mediux returned invalid show set data",
			map[string]any{
				"librarySection": librarySection,
				"tmdbID":         tmdbID,
				"setID":          setID,
			})
		return PosterSet{}, *actionShowSetValidation.Error
	}
	actionShowSetValidation.Complete()

	logAction.AppendResult("set_id", showSet.ID)
	logAction.AppendResult("set_name", showSet.SetTitle)
	logAction.AppendResult("date_created", showSet.DateCreated)
	logAction.AppendResult("date_updated", showSet.DateUpdated)

	// Process the response and return the poster sets
	posterSets := processShowResponse(ctx, librarySection, tmdbID, showSet.Show)

	if len(posterSets) == 0 {
		logAction.SetError("No Poster Sets Found", "No poster sets were found for the given show set",
			map[string]any{
				"librarySection": librarySection,
				"tmdbID":         tmdbID,
				"setID":          setID,
			})
		return PosterSet{}, *logAction.Error
	}

	posterSet := posterSets[0]
	posterSet.ID = showSet.ID
	posterSet.Title = showSet.SetTitle
	posterSet.User.Name = showSet.UserCreated.Username
	posterSet.Type = "show"
	posterSet.DateCreated = showSet.DateCreated
	posterSet.DateUpdated = showSet.DateUpdated

	return posterSet, logging.LogErrorInfo{}
}

func Mediux_FetchMovieSetByID(ctx context.Context, librarySection, tmdbID, setID string) (PosterSet, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetching Movie Set by ID from Mediux", logging.LevelTrace)
	defer logAction.Complete()

	// If cache is empty, return false
	if Global_Cache_LibraryStore.IsEmpty() {
		logAction.SetError("Cache is empty", "Cannot fetch movie set by ID because the library cache is empty", nil)
		return PosterSet{}, *logAction.Error
	}

	requestBody := Mediux_GenerateMovieSetByIDRequestBody(setID)

	// Send the GraphQL request
	resp, Err := Mediux_SendGraphQLRequest(ctx, requestBody)
	if Err.Message != "" {
		Err.Detail["librarySection"] = librarySection
		Err.Detail["tmdbID"] = tmdbID
		Err.Detail["setID"] = setID
		return PosterSet{}, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxMovieSetResponse
	Err = DecodeJSONBody(ctx, resp.Body(), &responseBody, "MediuxMovieSetResponse")
	if Err.Message != "" {
		return PosterSet{}, Err
	}

	movieSet := responseBody.Data.MovieSetID
	logAction.AppendResult("set_id", movieSet.ID)
	logAction.AppendResult("set_name", movieSet.SetTitle)
	logAction.AppendResult("date_created", movieSet.DateCreated)
	logAction.AppendResult("date_updated", movieSet.DateUpdated)

	// Process the response and return the poster sets
	posterSets := processMovieSetPostersAndBackdrops(ctx, librarySection, tmdbID, movieSet.Movie)
	if len(posterSets) == 0 {
		logAction.SetError("No Poster Sets Found", "No poster sets were found for the given movie set",
			map[string]any{
				"librarySection": librarySection,
				"tmdbID":         tmdbID,
				"setID":          setID,
			})
		return PosterSet{}, *logAction.Error
	}

	posterSet := posterSets[0]
	posterSet.ID = movieSet.ID
	posterSet.Title = movieSet.SetTitle
	posterSet.User.Name = movieSet.UserCreated.Username
	posterSet.Type = "movie"
	posterSet.DateCreated = movieSet.DateCreated
	posterSet.DateUpdated = movieSet.DateUpdated
	posterSet.Status = movieSet.Movie.Status

	return posterSet, logging.LogErrorInfo{}
}

func Mediux_FetchCollectionSetByID(ctx context.Context, librarySection, tmdbID, setID string) (PosterSet, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Fetching Collection Set by ID from Mediux", logging.LevelTrace)
	defer logAction.Complete()

	// If cache is empty, return false
	if Global_Cache_LibraryStore.IsEmpty() {
		logAction.SetError("Cache is empty", "Cannot fetch collection set by ID because the library cache is empty", nil)
		return PosterSet{}, *logAction.Error
	}

	requestBody := Mediux_GenerateCollectionSetByIDRequestBody(setID, tmdbID)

	// Send the GraphQL request
	resp, Err := Mediux_SendGraphQLRequest(ctx, requestBody)
	if Err.Message != "" {
		Err.Detail["librarySection"] = librarySection
		Err.Detail["tmdbID"] = tmdbID
		Err.Detail["setID"] = setID
		return PosterSet{}, Err
	}

	// Parse the response body into the appropriate struct based on itemType
	var responseBody MediuxCollectionSetResponse
	Err = DecodeJSONBody(ctx, resp.Body(), &responseBody, "MediuxCollectionSetResponse")
	if Err.Message != "" {
		return PosterSet{}, Err
	}

	collectionSet := responseBody.Data.CollectionSetID
	logAction.AppendResult("set_id", collectionSet.ID)
	logAction.AppendResult("set_name", collectionSet.SetTitle)
	logAction.AppendResult("date_created", collectionSet.DateCreated)
	logAction.AppendResult("date_updated", collectionSet.DateUpdated)

	if collectionSet.ID == "" {
		logAction.SetError("Invalid Collection Set Data", "Mediux returned invalid collection set data",
			map[string]any{
				"librarySection": librarySection,
				"tmdbID":         tmdbID,
				"setID":          setID,
			})
		return PosterSet{}, *logAction.Error
	}

	// Process the response and return the poster sets
	posterSets := processMovieCollection(ctx, tmdbID, librarySection, tmdbID, collectionSet.Collection)
	if len(posterSets) == 0 {
		logAction.SetError("No Poster Sets Found", "No poster sets were found for the given collection set",
			map[string]any{
				"librarySection": librarySection,
				"tmdbID":         tmdbID,
				"setID":          setID,
			})
		return PosterSet{}, *logAction.Error
	}

	posterSet := posterSets[0]
	posterSet.ID = collectionSet.ID
	posterSet.Title = collectionSet.SetTitle
	posterSet.User.Name = collectionSet.UserCreated.Username
	posterSet.Type = "collection"
	posterSet.DateCreated = collectionSet.DateCreated
	posterSet.DateUpdated = collectionSet.DateUpdated

	return posterSet, logging.LogErrorInfo{}
}
