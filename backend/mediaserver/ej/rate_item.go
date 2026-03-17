package ej

import (
	"aura/logging"
	"aura/models"
	"context"
)

// Not supported for Emby/Jellyfin, this is just a placeholder function to satisfy the MediaServerInterface
func (c *EJ) RateMediaItem(ctx context.Context, item *models.MediaItem, rating float64) (Err logging.LogErrorInfo) {
	return logging.LogErrorInfo{}
}
