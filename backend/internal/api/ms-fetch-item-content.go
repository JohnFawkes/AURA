package api

import (
	"aura/internal/logging"
)

func (p *PlexServer) FetchItemContent(ratingKey string, sectionTitle string) (MediaItem, logging.StandardError) {
	// Fetch the item content from Plex
	itemInfo, Err := Plex_FetchItemContent(ratingKey)
	if Err.Message != "" {
		return itemInfo, Err
	}
	return itemInfo, logging.StandardError{}
}

func (e *EmbyJellyServer) FetchItemContent(ratingKey string, sectionTitle string) (MediaItem, logging.StandardError) {
	// Fetch the item content from Emby/Jellyfin
	itemInfo, Err := EJ_FetchItemContent(ratingKey, sectionTitle)
	if Err.Message != "" {
		return itemInfo, Err
	}
	return itemInfo, logging.StandardError{}
}

func CallFetchItemContent(ratingKey string, sectionTitle string) (MediaItem, logging.StandardError) {

	mediaServer, Err := GetMediaServerInterface(Config_MediaServer{})
	if Err.Message != "" {
		return MediaItem{}, Err
	}

	return mediaServer.FetchItemContent(ratingKey, sectionTitle)
}
