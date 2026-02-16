package plex

import (
	"aura/config"
	"aura/logging"
	"context"
)

type Plex struct {
	Config config.Config_MediaServer
}

func (p *Plex) GetAdminUser(ctx context.Context, msConfig config.Config_MediaServer) (userID string, Err logging.LogErrorInfo) {
	// Plex does not require client initialization like Emby/Jellyfin
	return "", logging.LogErrorInfo{}
}
