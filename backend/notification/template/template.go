package notificationtmpl

import "sort"

const (
	EventAppStartup       = "app_startup"
	EventTestNotification = "test_notification"
)

const (
	VarAppName         = "AppName"
	VarAppVersion      = "AppVersion"
	VarAppPort         = "AppPort"
	VarMediaServerName = "MediaServerName"
	VarMediaServerType = "MediaServerType"
	VarTimestamp       = "Timestamp"
)

var AllowedByEvent = map[string][]string{
	EventAppStartup: {
		VarAppName, VarAppVersion, VarAppPort,
		VarMediaServerName, VarMediaServerType, VarTimestamp,
	},
	EventTestNotification: {
		VarMediaServerName, VarMediaServerType, VarTimestamp,
	},
}

func WrappedAllowed(event string) []string {
	raw := AllowedByEvent[event]
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		out = append(out, "{{"+v+"}}")
	}
	sort.Strings(out)
	return out
}
