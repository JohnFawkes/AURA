package api

import (
	"aura/internal/masking"
)

func (config *Config) Sanitize() Config {

	if config == nil {
		return Config{}
	}

	// Start with a shallow copy of the root struct.
	c := *config

	// Mask top-level sensitive fields (ensure these are value fields, not shared pointers).
	//c.Auth.Password = MaskToken(c.Auth.Password)
	c.Mediux.Token = masking.Masking_Token(c.Mediux.Token)
	c.TMDB.APIKey = masking.Masking_Token(c.TMDB.APIKey)
	c.MediaServer.Token = masking.Masking_Token(c.MediaServer.Token)

	// Deep copy notifications.providers slice and nested pointer
	if len(config.Notifications.Providers) > 0 {
		c.Notifications.Providers = make([]Config_Notification_Provider, len(config.Notifications.Providers))
		for i, p := range config.Notifications.Providers {
			cp := p // copy struct
			if p.Discord != nil {
				cp.Discord = &Config_Notification_Discord{
					Webhook: masking.Masking_WebhookURL(p.Discord.Webhook),
				}
			}
			if p.Pushover != nil {
				cp.Pushover = &Config_Notification_Pushover{
					Token:   masking.Masking_Token(p.Pushover.Token),
					UserKey: masking.Masking_Token(p.Pushover.UserKey),
				}
			}
			if p.Gotify != nil {
				cp.Gotify = &Config_Notification_Gotify{
					URL:   p.Gotify.URL, // URL is not sensitive
					Token: masking.Masking_Token(p.Gotify.Token),
				}
			}
			c.Notifications.Providers[i] = cp
		}
	}

	// Deep copy SonarrRadarr slice
	if len(config.SonarrRadarr.Applications) > 0 {
		c.SonarrRadarr.Applications = make([]Config_SonarrRadarrApp, len(config.SonarrRadarr.Applications))
		for i, app := range config.SonarrRadarr.Applications {
			c.SonarrRadarr.Applications[i] = Config_SonarrRadarrApp{
				Type:    app.Type,
				Library: app.Library,
				URL:     app.URL,
				APIKey:  masking.Masking_Token(app.APIKey),
			}
		}
	}

	return c
}
