package mediaserver_shared

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/mediux"
	"aura/internal/modals"
	"aura/internal/server/emby_jellyfin"
	"aura/internal/server/plex"
	"aura/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type MediaServer interface {

	// Get Status of the Media Server
	GetMediaServerStatus() (string, logging.StandardError)

	// Get the library section info
	FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.StandardError)

	// Get the library section items
	FetchLibrarySectionItems(section modals.LibrarySection, sectionStartIndex string) ([]modals.MediaItem, int, logging.StandardError)

	// Get an item's content by Rating Key/ID
	FetchItemContent(ratingKey string, sectionTitle string) (modals.MediaItem, logging.StandardError)

	// Get an image from the media server
	FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.StandardError)

	// Use the set to update the item on the media server
	DownloadAndUpdatePosters(mediaItem modals.MediaItem, file modals.PosterFile) logging.StandardError

	// Use the TMDB ID, type, title and library section to search for the item on the media server
	SearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection string) (string, logging.StandardError)
}

type PlexServer struct{}
type EmbyJellyServer struct{}

func (p *PlexServer) GetMediaServerStatus() (string, logging.StandardError) {
	// Get the status of the Plex server
	version, Err := plex.GetMediaServerStatus()
	if Err.Message != "" {
		return "", Err
	}
	return version, logging.StandardError{}
}

func (e *EmbyJellyServer) GetMediaServerStatus() (string, logging.StandardError) {
	//Get the status of the Emby/Jellyfin server
	version, Err := emby_jellyfin.GetMediaServerStatus()
	if Err.Message != "" {
		return "", Err
	}
	return version, logging.StandardError{}
}

func (p *PlexServer) FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.StandardError) {
	// Fetch the library section from Plex
	found, Err := plex.FetchLibrarySectionInfo(library)
	if Err.Message != "" {
		return false, Err
	}
	return found, logging.StandardError{}
}

func (e *EmbyJellyServer) FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.StandardError) {
	// Fetch the library section from Emby/Jellyfin
	found, Err := emby_jellyfin.FetchLibrarySectionInfo(library)
	if Err.Message != "" {
		return false, Err
	}
	return found, logging.StandardError{}
}

func (p *PlexServer) FetchLibrarySectionItems(section modals.LibrarySection, sectionStartIndex string) ([]modals.MediaItem, int, logging.StandardError) {
	// Fetch the section content from Plex
	mediaItems, totalSize, Err := plex.FetchLibrarySectionItems(section, sectionStartIndex, "500")
	if Err.Message != "" {
		return nil, 0, Err
	}
	return mediaItems, totalSize, logging.StandardError{}
}

func (e *EmbyJellyServer) FetchLibrarySectionItems(section modals.LibrarySection, sectionStartIndex string) ([]modals.MediaItem, int, logging.StandardError) {
	// Fetch the section content from Emby/Jellyfin
	mediaItems, totalSize, Err := emby_jellyfin.FetchLibrarySectionItems(section, sectionStartIndex, "500")
	if Err.Message != "" {
		return nil, 0, Err
	}
	return mediaItems, totalSize, logging.StandardError{}
}

func (p *PlexServer) FetchItemContent(ratingKey string, sectionTitle string) (modals.MediaItem, logging.StandardError) {
	// Fetch the item content from Plex
	itemInfo, Err := plex.FetchItemContent(ratingKey)
	if Err.Message != "" {
		return itemInfo, Err
	}
	return itemInfo, logging.StandardError{}
}

func (e *EmbyJellyServer) FetchItemContent(ratingKey string, sectionTitle string) (modals.MediaItem, logging.StandardError) {
	// Fetch the item content from Emby/Jellyfin
	itemInfo, Err := emby_jellyfin.FetchItemContent(ratingKey, sectionTitle)
	if Err.Message != "" {
		return itemInfo, Err
	}
	return itemInfo, logging.StandardError{}
}

func (p *PlexServer) FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.StandardError) {
	// Fetch the image from Plex
	imageData, Err := plex.FetchImageFromMediaServer(ratingKey, imageType)
	if Err.Message != "" {
		return nil, Err
	}
	return imageData, logging.StandardError{}
}

func (e *EmbyJellyServer) FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.StandardError) {
	// Fetch the image from Emby/Jellyfin
	imageData, Err := emby_jellyfin.FetchImageFromMediaServer(ratingKey, imageType)
	if Err.Message != "" {
		return nil, Err
	}
	return imageData, logging.StandardError{}
}

func (p *PlexServer) DownloadAndUpdatePosters(mediaItem modals.MediaItem, file modals.PosterFile) logging.StandardError {
	// Download and update the item on Plex
	Err := plex.DownloadAndUpdatePosters(mediaItem, file)
	if Err.Message != "" {
		return Err
	}
	return logging.StandardError{}
}

func (e *EmbyJellyServer) DownloadAndUpdatePosters(mediaItem modals.MediaItem, file modals.PosterFile) logging.StandardError {
	// Download and update the item on Emby/Jellyfin
	Err := emby_jellyfin.DownloadAndUpdatePosters(mediaItem, file)
	if Err.Message != "" {
		return Err
	}
	return logging.StandardError{}
}

func (p *PlexServer) SearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection string) (string, logging.StandardError) {
	// Search for the item on Plex
	ratingKey, Err := mediux.PlexSearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection)
	if Err.Message != "" {
		return ratingKey, Err
	}
	return ratingKey, logging.StandardError{}
}

func (e *EmbyJellyServer) SearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection string) (string, logging.StandardError) {
	// Search for the item on Emby/Jellyfin
	ratingKey, Err := mediux.EmbyJellySearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection)
	if Err.Message != "" {
		return ratingKey, Err
	}
	return ratingKey, logging.StandardError{}
}

func InitUserID() logging.StandardError {
	if config.Global.MediaServer.Type == "Plex" {
		return logging.StandardError{}
	}

	logging.LOG.Debug(fmt.Sprintf("Initializing UserID for %s Media Server", config.Global.MediaServer.Type))

	Err := logging.NewStandardError()

	// Parse the base URL
	baseURL, err := url.Parse(config.Global.MediaServer.URL)
	if err != nil {

		Err.Message = "Failed to parse base URL"
		Err.HelpText = "Ensure the Media Server URL is correctly configured in the settings."
		Err.Details = fmt.Sprintf("Base URL: %s", config.Global.MediaServer.URL)
		return Err
	}
	// Construct the full URL by appending the path
	baseURL.Path = path.Join(baseURL.Path, "Users")
	url := baseURL.String()

	// Make a GET request to the Emby server
	response, body, Err := utils.MakeHTTPRequest(url, http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	var responseSection modals.EmbyJellyUserIDResponse
	err = json.Unmarshal(body, &responseSection)
	if err != nil {

		Err.Message = "Failed to parse Emby/Jellyfin user ID response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = fmt.Sprintf("Error: %s", err.Error())
		return Err
	}

	// Find the first Admin user ID
	for _, user := range responseSection {
		if user.Policy.IsAdministrator {
			config.Global.MediaServer.UserID = user.ID
			maskedUserID := fmt.Sprintf("****%s", user.ID[len(user.ID)-7:])
			logging.LOG.Debug(fmt.Sprintf("Found Admin user ID: %s", maskedUserID))
			return logging.StandardError{}
		}
	}

	// If no Admin user is found, return an error

	Err.Message = "No Admin user found in Emby/Jellyfin user list"
	Err.HelpText = "Ensure that there is at least one user with Admin privileges in the Emby/Jellyfin server."
	Err.Details = "No Admin user found in the Emby/Jellyfin user list"
	return Err

}
