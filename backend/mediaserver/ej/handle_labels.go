package ej

import (
	"aura/logging"
	"aura/models"
	"context"
)

func (ej *EJ) AddLabelToMediaItem(ctx context.Context, item models.MediaItem, selectedTypes models.SelectedTypes) (Err logging.LogErrorInfo) {
	// Emby does not support labels, so this function will simply return without doing anything
	return logging.LogErrorInfo{}
}
