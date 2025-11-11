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
	TMDBID       string      `json:"TMDBID,omitempty"`
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

	var collectionItems []CollectionItem

	// Construct the URL for the Emby/Jellyfin server API request
	u, err := url.Parse(Global_Config.MediaServer.URL)
	if err != nil {
		logAction.SetError("Failed to parse URL", err.Error(), nil)
		return collectionItems, *logAction.Error
	}
	u.Path = path.Join(u.Path, "Users", Global_Config.MediaServer.UserID, "Items")
	query := u.Query()
	query.Add("Recursive", "true")
	query.Add("SortBy", "Name")
	query.Add("SortOrder", "Ascending")
	query.Add("IncludeItemTypes", "BoxSet")
	query.Add("ParentId", librarySection.ID)

	u.RawQuery = query.Encode()
	URL := u.String()

	// Make Auth Headers for Request
	headers := MakeAuthHeader("X-Emby-Token", Global_Config.MediaServer.Token)

	// Make the HTTP request to Emby/Jellyfin
	httpResp, respBody, logErr := MakeHTTPRequest(ctx, URL, http.MethodGet, headers, 60, nil, Global_Config.MediaServer.Type)
	if logErr.Message != "" {
		return collectionItems, logErr
	}
	defer httpResp.Body.Close()

	var responseSection EmbyJellyLibraryItemsResponse
	logErr = DecodeJSONBody(ctx, respBody, &responseSection, "EmbyJellyLibraryItemsResponse")
	if logErr.Message != "" {
		return collectionItems, logErr
	}

	// Now we loop through the items and build out the collection items
	// Note: Emby/Jellyfin does not have collections in the same way Plex does.
	// So we will need to loop through each boxset and get the details within.
	// This will tell us if the Boxset has a TMDB ID associated with it.

	for _, item := range responseSection.Items {
		if item.Type != "BoxSet" || item.ID == "" {
			continue
		}
		subAction := logAction.AddSubAction(fmt.Sprintf("Processing BoxSet '%s' (ID: %s)", item.Name, item.ID), logging.LevelTrace)

		// Construct the URL for the Emby/Jellyfin server API request
		u, err := url.Parse(Global_Config.MediaServer.URL)
		if err != nil {
			subAction.AppendWarning("message", fmt.Sprintf("Failed to parse URL: %s", err.Error()))
			continue
		}
		u.Path = path.Join(u.Path, "Users", Global_Config.MediaServer.UserID, "Items", item.ID)

		// Make the API request to Emby/Jellyfin
		itemHTTPResp, itemRespBody, itemLogErr := MakeHTTPRequest(ctx, u.String(), http.MethodGet, headers, 60, nil, Global_Config.MediaServer.Type)
		if itemLogErr.Message != "" {
			subAction.AppendWarning("message", fmt.Sprintf("Failed to fetch BoxSet details: %s", itemLogErr.Message))
			continue
		}
		defer itemHTTPResp.Body.Close()

		// Parse the response body into an EmbyJellyItemContentResponse struct
		var ejResponse EmbyJellyItemContentResponse
		logErr = DecodeJSONBody(ctx, itemRespBody, &ejResponse, "EmbyJellyItemContentResponse")
		if logErr.Message != "" {
			subAction.AppendWarning("message", fmt.Sprintf("Failed to decode BoxSet details: %s", logErr.Message))
			continue
		}

		// If item doesn't have a TMDB ID, skip it
		if ejResponse.ProviderIds.Tmdb == "" {
			subAction.AppendWarning("message", "boxset does not have a TMDB ID; skipping")
			continue
		}

		var collectionItem CollectionItem
		collectionItem.RatingKey = ejResponse.ID
		collectionItem.TMDBID = ejResponse.ProviderIds.Tmdb
		collectionItem.Title = ejResponse.Name
		collectionItem.Summary = ejResponse.Overview
		collectionItem.ChildCount = ejResponse.ChildCount
		collectionItem.LibraryTitle = librarySection.Title
		collectionItems = append(collectionItems, collectionItem)

		subAction.Complete()
	}

	return collectionItems, logging.LogErrorInfo{}
}
