package api

import (
	"aura/internal/logging"
	"context"
)

func (p *PlexServer) FetchItemContent(ctx context.Context, ratingKey string, sectionTitle string) (MediaItem, logging.LogErrorInfo) {
	return Plex_FetchItemContent(ctx, ratingKey)
}

func (e *EmbyJellyServer) FetchItemContent(ctx context.Context, ratingKey string, sectionTitle string) (MediaItem, logging.LogErrorInfo) {
	return EJ_FetchItemContent(ctx, ratingKey, sectionTitle)
}

//////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////

func CallFetchItemContent(ctx context.Context, ratingKey string, sectionTitle string) (MediaItem, logging.LogErrorInfo) {
	mediaServer, _, Err := NewMediaServerInterface(ctx, Config_MediaServer{})
	if Err.Message != "" {
		return MediaItem{}, Err
	}

	return mediaServer.FetchItemContent(ctx, ratingKey, sectionTitle)
}
