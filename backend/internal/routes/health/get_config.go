package health

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils"
	"fmt"
	"net/http"
	"time"
)

func GetConfig(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	logging.LOG.Trace(r.URL.Path)

	safeConfigData := SanitizedCopy()

	logging.LOG.Trace(fmt.Sprintf("Safe Config Data: %+v", safeConfigData))

	// (If Logging.File is a pointer or needs refresh, set after clone)
	safeConfigData.Logging.File = logging.GetTodayLogFile()

	utils.SendJsonResponse(w, http.StatusOK, utils.JSONResponse{
		Status:  "success",
		Elapsed: utils.ElapsedTime(startTime),
		Data:    safeConfigData,
	})
}

// SanitizedCopy returns a deep-copied, masked version of the global config safe for exposure.
func SanitizedCopy() modals.Config {
	if config.Global == nil {
		logging.LOG.Warn("SanitizedCopy: Global config is nil")
		return modals.Config{}
	}

	// Start with a shallow copy of the root struct.
	c := *config.Global

	// Mask top-level sensitive fields (ensure these are value fields, not shared pointers).
	c.Auth.Password = config.MaskToken(c.Auth.Password)
	c.Mediux.Token = config.MaskToken(c.Mediux.Token)
	c.TMDB.ApiKey = config.MaskToken(c.TMDB.ApiKey)
	c.MediaServer.Token = config.MaskToken(c.MediaServer.Token)

	// Deep copy notifications.providers slice and nested pointer structs.
	if len(config.Global.Notifications.Providers) > 0 {
		c.Notifications.Providers = make([]modals.Config_Notification_Providers, len(config.Global.Notifications.Providers))
		for i, p := range config.Global.Notifications.Providers {
			cp := p // copy struct
			if p.Discord != nil {
				cp.Discord = &modals.Config_Notification_Discord{
					Webhook: config.MaskWebhookURL(p.Discord.Webhook),
				}
			}
			if p.Pushover != nil {
				cp.Pushover = &modals.Config_Notification_Pushover{
					Token:   config.MaskToken(p.Pushover.Token),
					UserKey: config.MaskToken(p.Pushover.UserKey),
				}
			}
			c.Notifications.Providers[i] = cp
		}
	}

	return c
}
