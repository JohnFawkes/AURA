package api

import (
	"aura/internal/logging"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type Interface_MediaServer interface {

	// Get Status of the Media Server
	GetMediaServerStatus(msConfig Config_MediaServer) (string, logging.StandardError)

	// Get the library section info
	FetchLibrarySectionInfo(library *Config_MediaServerLibrary) (bool, logging.StandardError)

	// Get the library section options
	FetchLibrarySectionOptions() ([]string, logging.StandardError)

	// Get the library section items
	FetchLibrarySectionItems(section LibrarySection, sectionStartIndex string) ([]MediaItem, int, logging.StandardError)

	// Get an item's content by Rating Key/ID
	FetchItemContent(ratingKey string, sectionTitle string) (MediaItem, logging.StandardError)

	// Get an image from the media server
	FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.StandardError)

	// Use the set to update the item on the media server
	DownloadAndUpdatePosters(mediaItem MediaItem, file PosterFile) logging.StandardError

	// Use the TMDB ID, type, title and library section to search for the item on the media server
	//SearchForItemAndGetRatingKey(tmdbID, itemType, itemTitle, librarySection string) (string, logging.StandardError)
}

type PlexServer struct{}
type EmbyJellyServer struct{}

func GetMediaServerInterface(msConfig Config_MediaServer) (Interface_MediaServer, logging.StandardError) {
	if msConfig.Type == "" && msConfig.URL == "" && msConfig.Token == "" {
		msConfig = Global_Config.MediaServer
	}
	var interfaceMS Interface_MediaServer
	switch msConfig.Type {
	case "Plex":
		interfaceMS = &PlexServer{}
	case "Emby", "Jellyfin":
		interfaceMS = &EmbyJellyServer{}
	default:
		Err := logging.NewStandardError()
		Err.Message = "Unsupported Media Server Type"
		Err.HelpText = "Ensure the Media Server Type is set to either 'Plex', 'Emby', or 'Jellyfin' in the configuration."
		return nil, Err
	}

	return interfaceMS, logging.StandardError{}
}

func MediaServer_Init(msConfig Config_MediaServer) logging.StandardError {
	if msConfig.Type == "Plex" {
		return CheckPlexConnection(msConfig)
	}

	logging.LOG.Debug(fmt.Sprintf("Initializing UserID for %s Media Server", msConfig.Type))

	Err := logging.NewStandardError()

	// Parse the base URL
	baseURL, err := url.Parse(msConfig.URL)
	if err != nil {
		Err.Message = "Failed to parse base URL"
		Err.HelpText = "Ensure the Media Server URL is correctly configured in the settings."
		Err.Details = map[string]any{
			"error":   err.Error(),
			"request": baseURL.String(),
		}
		return Err
	}
	// Construct the full URL by appending the path
	baseURL.Path = path.Join(baseURL.Path, "Users")
	url := baseURL.String()

	// Make a GET request to the Jellyfin/Emby server
	response, body, Err := MakeHTTPRequest(url, http.MethodGet, nil, 60, nil, "MediaServer")
	if Err.Message != "" {
		return Err
	}
	defer response.Body.Close()

	var responseSection EmbyJellyUserIDResponse
	err = json.Unmarshal(body, &responseSection)
	if err != nil {
		Err.Message = "Failed to parse Emby/Jellyfin user ID response"
		Err.HelpText = "Ensure the Emby/Jellyfin server is returning a valid JSON response."
		Err.Details = map[string]any{
			"error":   err.Error(),
			"request": baseURL.String(),
		}
		return Err
	}

	// Find the first Admin user ID
	for _, user := range responseSection {
		if user.Policy.IsAdministrator {
			Global_Config.MediaServer.UserID = user.ID
			maskedUserID := fmt.Sprintf("****%s", user.ID[len(user.ID)-7:])
			logging.LOG.Debug(fmt.Sprintf("Found Admin user ID: %s", maskedUserID))
			return logging.StandardError{}
		}
	}

	// If no Admin user is found, return an error
	Err.Message = "No Admin user found in Emby/Jellyfin user list"
	Err.HelpText = "Ensure that there is at least one user with Admin privileges in the Emby/Jellyfin server."
	Err.Details = map[string]any{
		"request": baseURL.String(),
	}
	return Err

}

func CheckPlexConnection(msConfig Config_MediaServer) logging.StandardError {
	if msConfig.Type != "Plex" {
		return logging.StandardError{}
	}

	logging.LOG.Debug("Checking connection to Plex Media Server")

	Err := logging.NewStandardError()

	version, Err := Plex_GetMediaServerStatus(msConfig)
	if Err.Message != "" {
		return Err
	}

	logging.LOG.Info(fmt.Sprintf("Successfully connected to Plex Media Server (Version: %s)", version))

	return logging.StandardError{}
}
