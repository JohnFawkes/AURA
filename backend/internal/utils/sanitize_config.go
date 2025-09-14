package utils

import (
	"aura/internal/config"
	"aura/internal/logging"
	"aura/internal/modals"
	"aura/internal/utils/masking"
)

func SanitizedCopy() modals.Config {
	if config.Global == nil {
		logging.LOG.Warn("SanitizedCopy: Global config is nil")
		return modals.Config{}
	}

	// Start with a shallow copy of the root struct.
	c := *config.Global

	// Mask top-level sensitive fields (ensure these are value fields, not shared pointers).
	//c.Auth.Password = MaskToken(c.Auth.Password)
	c.Mediux.Token = masking.MaskToken(c.Mediux.Token)
	c.TMDB.ApiKey = masking.MaskToken(c.TMDB.ApiKey)
	c.MediaServer.Token = masking.MaskToken(c.MediaServer.Token)

	// Deep copy notifications.providers slice and nested pointer structs.
	if len(config.Global.Notifications.Providers) > 0 {
		c.Notifications.Providers = make([]modals.Config_Notification_Providers, len(config.Global.Notifications.Providers))
		for i, p := range config.Global.Notifications.Providers {
			cp := p // copy struct
			if p.Discord != nil {
				cp.Discord = &modals.Config_Notification_Discord{
					Webhook: masking.MaskWebhookURL(p.Discord.Webhook),
				}
			}
			if p.Pushover != nil {
				cp.Pushover = &modals.Config_Notification_Pushover{
					Token:   masking.MaskToken(p.Pushover.Token),
					UserKey: masking.MaskToken(p.Pushover.UserKey),
				}
			}
			if p.Gotify != nil {
				cp.Gotify = &modals.Config_Notification_Gotify{
					URL:   p.Gotify.URL, // URL is not sensitive
					Token: masking.MaskToken(p.Gotify.Token),
				}
			}
			c.Notifications.Providers[i] = cp
		}
	}

	return c
}
