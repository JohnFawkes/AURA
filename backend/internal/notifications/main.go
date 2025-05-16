package notifications

import "poster-setter/internal/config"

// Valid NotificationProviders is a list of valid notification providers
var ValidNotificationProviders = []string{
	"Discord",
}

func validNotificationProvider() bool {
	for _, provider := range ValidNotificationProviders {
		if config.Global.Notification.Provider == provider {
			return true
		}
	}
	return false
}
