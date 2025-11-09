package api

import (
	"aura/internal/logging"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type CollectionItem struct {
	RatingKey    string      `json:"RatingKey"`
	Title        string      `json:"Title"`
	Summary      string      `json:"Summary,omitempty"`
	ChildCount   int         `json:"ChildCount"`
	MediaItems   []MediaItem `json:"MediaItems"`
	LibraryTitle string      `json:"LibraryTitle,omitempty"`
}

func (p *PlexServer) FetchMovieCollectionItems(ctx context.Context, librarySection LibrarySection) ([]CollectionItem, logging.LogErrorInfo) {
	return Plex_FetchMovieCollectionItems(ctx, librarySection)
}

func (e *EmbyJellyServer) FetchMovieCollectionItems(ctx context.Context, librarySection LibrarySection) ([]CollectionItem, logging.LogErrorInfo) {
	return EJ_FetchMovieCollectionItems(ctx, librarySection)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallFetchMovieCollectionItems(ctx context.Context, librarySection LibrarySection) ([]CollectionItem, logging.LogErrorInfo) {
	mediaServer, _, Err := NewMediaServerInterface(ctx, Config_MediaServer{})
	if Err.Message != "" {
		return []CollectionItem{}, Err
	}
	return mediaServer.FetchMovieCollectionItems(ctx, librarySection)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func Plex_FetchMovieCollectionItems(ctx context.Context, librarySection LibrarySection) ([]CollectionItem, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx,
		fmt.Sprintf("Getting all collections for '%s' (ID: %s)", librarySection.Title, librarySection.ID), logging.LevelDebug)
	defer logAction.Complete()

	var collectionItems []CollectionItem

	// Construct the URL for the Plex server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse URL", err.Error(), nil)
		return collectionItems, *logAction.Error
	}
	u.Path = path.Join(u.Path, "library", "sections", librarySection.ID, "collections")
	query := u.Query()
	query.Set("includeCollections", "1")
	u.RawQuery = query.Encode()
	URL := u.String()

	// Make the Auth Headers for Request
	headers := MakeAuthHeader("X-Plex-Token", Global_Config.MediaServer.Token)

	// Make the HTTP request to Plex
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, "Plex")
	if logErr.Message != "" {
		return collectionItems, logErr
	}
	defer httpResp.Body.Close()

	// Check the response status code
	if httpResp.StatusCode != 200 {
		logAction.SetError("Plex server returned non-200 status", fmt.Sprintf("Status Code: %d", httpResp.StatusCode), nil)
		return collectionItems, *logAction.Error
	}

	// Decode the response body
	var plexResponse PlexLibraryItemsWrapper
	logErr = DecodeJSONBody(ctx, respBody, &plexResponse, "PlexLibraryItemsWrapper")
	if logErr.Message != "" {
		return collectionItems, logErr
	}

	for _, collection := range plexResponse.MediaContainer.Metadata {
		var collectionItem CollectionItem
		if collection.Type != "collection" || collection.ChildCount < 1 {
			continue
		}

		collectionItem.RatingKey = collection.RatingKey
		collectionItem.Title = collection.Title
		collectionItem.Summary = collection.Summary
		collectionItem.ChildCount = collection.ChildCount
		collectionItem.LibraryTitle = librarySection.Title
		collectionItems = append(collectionItems, collectionItem)
	}

	return collectionItems, logging.LogErrorInfo{}
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func EJ_FetchMovieCollectionItems(ctx context.Context, librarySection LibrarySection) ([]CollectionItem, logging.LogErrorInfo) {
	ctx, logAction := logging.AddSubActionToContext(ctx,
		fmt.Sprintf("Getting all collections for '%s' (ID: %s)", librarySection.Title, librarySection.ID), logging.LevelDebug)
	defer logAction.Complete()

	logAction.SetError(fmt.Sprintf("Function not implemented for %s 🚧", Global_Config.MediaServer.Type), "Hopefully coming soon! 🦄", nil)

	var collectionItems []CollectionItem
	return collectionItems, logging.LogErrorInfo{}
}
