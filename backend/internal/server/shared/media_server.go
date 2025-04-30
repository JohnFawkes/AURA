package mediaserver_shared

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"poster-setter/internal/config"
	"poster-setter/internal/logging"
	"poster-setter/internal/modals"
	"poster-setter/internal/server/emby_jellyfin"
	"poster-setter/internal/server/plex"
	"poster-setter/internal/utils"
)

type MediaServer interface {
	// Get the library section info
	FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.ErrorLog)

	// Get the library section items
	FetchLibrarySectionItems(sectionID, sectionTitle string) ([]modals.MediaItem, logging.ErrorLog)

	// Get an item's content by Rating Key/ID
	FetchItemContent(ratingKey string, sectionTitle string) (modals.MediaItem, logging.ErrorLog)

	// Get an image from the media server
	FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.ErrorLog)

	// Use the set to update the item on the media server
	DownloadAndUpdatePosters(mediaItem modals.MediaItem, file modals.PosterFile) logging.ErrorLog
}

type PlexServer struct{}
type EmbyJellyServer struct{}

func (p *PlexServer) FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.ErrorLog) {
	// Fetch the library section from Plex
	found, logErr := plex.FetchLibrarySectionInfo(library)
	if logErr.Err != nil {
		return false, logErr
	}
	return found, logging.ErrorLog{}
}

func (e *EmbyJellyServer) FetchLibrarySectionInfo(library *modals.Config_MediaServerLibrary) (bool, logging.ErrorLog) {
	// Fetch the library section from Emby/Jellyfin
	found, logErr := emby_jellyfin.FetchLibrarySectionInfo(library)
	if logErr.Err != nil {
		return false, logErr
	}
	return found, logging.ErrorLog{}
}

func (p *PlexServer) FetchLibrarySectionItems(sectionID, sectionTitle string) ([]modals.MediaItem, logging.ErrorLog) {
	// Fetch the section content from Plex
	mediaItems, logErr := plex.FetchLibrarySectionItems(sectionID)
	if logErr.Err != nil {
		return nil, logErr
	}
	return mediaItems, logging.ErrorLog{}
}

func (e *EmbyJellyServer) FetchLibrarySectionItems(sectionID, sectionTitle string) ([]modals.MediaItem, logging.ErrorLog) {
	// Fetch the section content from Emby/Jellyfin
	mediaItems, logErr := emby_jellyfin.FetchLibrarySectionItems(sectionID, sectionTitle)
	if logErr.Err != nil {
		return nil, logErr
	}
	return mediaItems, logging.ErrorLog{}
}

func (p *PlexServer) FetchItemContent(ratingKey string, sectionTitle string) (modals.MediaItem, logging.ErrorLog) {
	// Fetch the item content from Plex
	itemInfo, logErr := plex.FetchItemContent(ratingKey)
	if logErr.Err != nil {
		return itemInfo, logErr
	}
	return itemInfo, logging.ErrorLog{}
}

func (e *EmbyJellyServer) FetchItemContent(ratingKey string, sectionTitle string) (modals.MediaItem, logging.ErrorLog) {
	// Fetch the item content from Emby/Jellyfin
	itemInfo, logErr := emby_jellyfin.FetchItemContent(ratingKey, sectionTitle)
	if logErr.Err != nil {
		return itemInfo, logErr
	}
	return itemInfo, logging.ErrorLog{}
}

func (p *PlexServer) FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.ErrorLog) {
	// Fetch the image from Plex
	imageData, logErr := plex.FetchImageFromMediaServer(ratingKey, imageType)
	if logErr.Err != nil {
		return nil, logErr
	}
	return imageData, logging.ErrorLog{}
}

func (e *EmbyJellyServer) FetchImageFromMediaServer(ratingKey, imageType string) ([]byte, logging.ErrorLog) {
	// Fetch the image from Emby/Jellyfin
	imageData, logErr := emby_jellyfin.FetchImageFromMediaServer(ratingKey, imageType)
	if logErr.Err != nil {
		return nil, logErr
	}
	return imageData, logging.ErrorLog{}
}

func (p *PlexServer) DownloadAndUpdatePosters(mediaItem modals.MediaItem, file modals.PosterFile) logging.ErrorLog {
	// Download and update the item on Plex
	logErr := plex.DownloadAndUpdatePosters(mediaItem, file)
	if logErr.Err != nil {
		return logErr
	}
	return logging.ErrorLog{}
}

func (e *EmbyJellyServer) DownloadAndUpdatePosters(mediaItem modals.MediaItem, file modals.PosterFile) logging.ErrorLog {
	// Download and update the item on Emby/Jellyfin
	logErr := emby_jellyfin.DownloadAndUpdatePosters(mediaItem, file)
	if logErr.Err != nil {
		return logErr
	}
	return logging.ErrorLog{}
}

func InitUserID() logging.ErrorLog {

	if config.Global.MediaServer.Type == "Plex" {
		return logging.ErrorLog{}
	}

	logging.LOG.Trace("Initializing User ID for Emby/Jellyfin")
	// Parse the base URL
	baseURL, err := url.Parse(config.Global.MediaServer.URL)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{Message: "Invalid base URL"}}
	}
	// Construct the full URL by appending the path
	baseURL.Path = path.Join(baseURL.Path, "Users")
	url := baseURL.String()

	// Make a GET request to the Emby server
	response, body, logErr := utils.MakeHTTPRequest(url, http.MethodGet, nil, 60, nil, "MediaServer")
	if logErr.Err != nil {
		return logErr
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return logging.ErrorLog{Err: fmt.Errorf("received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type),
			Log: logging.Log{Message: fmt.Sprintf("Received status code '%d' from %s server", response.StatusCode, config.Global.MediaServer.Type)}}
	}

	var responseSection modals.EmbyJellyUserIDResponse
	err = json.Unmarshal(body, &responseSection)
	if err != nil {
		return logging.ErrorLog{Err: err, Log: logging.Log{Message: "Failed to parse JSON response"}}
	}

	// Find the first Admin user ID
	for _, user := range responseSection {
		if user.Policy.IsAdministrator {
			config.Global.MediaServer.UserID = user.ID
			maskedUserID := fmt.Sprintf("****%s", user.ID[len(user.ID)-7:])
			logging.LOG.Trace(fmt.Sprintf("Found Admin user ID: %s", maskedUserID))
			return logging.ErrorLog{}
		}
	}

	return logging.ErrorLog{Err: fmt.Errorf("no admin user found"),
		Log: logging.Log{Message: "No Admin user found"}}
}
