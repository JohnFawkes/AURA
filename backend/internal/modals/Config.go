package modals

// Config represents the main application configuration settings.
// It includes options for server setup, caching, data storage, logging,
// and integration with external services such as Plex, TMDB, and Mediux.
type Config struct {
	Auth                   Config_Auth          `yaml:"Auth"`                   // Authentication settings.
	Dev                    Config_Dev           `yaml:"Dev"`                    // Development mode settings.
	CacheImages            bool                 `yaml:"CacheImages"`            // Whether to cache images locally.
	SaveImageNextToContent bool                 `yaml:"SaveImageNextToContent"` // Whether to save images next to the associated content.
	Logging                Config_Logging       `yaml:"Logging"`                // Logging configuration settings.
	MediaServer            Config_MediaServer   `yaml:"MediaServer"`            // Media server integration settings.
	TMDB                   Config_TMDB          `yaml:"TMDB"`                   // TMDB (The Movie Database) integration settings.
	Mediux                 Config_Mediux        `yaml:"Mediux"`                 // Mediux integration settings.
	AutoDownload           Config_AutoDownload  `yaml:"AutoDownload"`           // Auto-download settings.
	Kometa                 Config_Kometa        `yaml:"Kometa"`                 // Kometa settings.
	Notifications          Config_Notifications `yaml:"Notifications"`          // Notification settings.
}

type Config_Auth struct {
	Enable   bool   `yaml:"Enable"`   // Whether to enable authentication.
	Password string `yaml:"Password"` // Password for authentication.
}

type Config_Dev struct {
	Enable    bool   `yaml:"Enable"`    // Whether to enable development mode.
	LocalPath string `yaml:"LocalPath"` // Local path for development mode.
}

// Config_Logging represents the logging configuration settings.
type Config_Logging struct {
	Level string `yaml:"Level"` // Logging level (e.g., TRACE, DEBUG, INFO, WARN, ERROR).
	File  string `yaml:"File"`  // File path for logging output.
}

// Config_MediaServerLibrary represents a single media server library configuration.
type Config_MediaServerLibrary struct {
	Name      string `yaml:"Name"`      // Name of the library.
	SectionID string `yaml:"SectionID"` // Unique identifier for the library section.
	Type      string `yaml:"Type"`      // Type of the library (e.g., movie, show). All other types are ignored.
}

// Config_MediaServer represents the configuration for media server integration.
type Config_MediaServer struct {
	Type                   string                      `yaml:"Type"`                   // Type of media server (e.g., plex, emby, jellyfin).
	URL                    string                      `yaml:"URL"`                    // Base URL of the media server. This is either the IP:Port or the domain name (e.g., plex.domain.com).
	Token                  string                      `yaml:"Token"`                  // Authentication token for accessing the media server.
	Libraries              []Config_MediaServerLibrary `yaml:"Libraries"`              // List of media server libraries to manage.
	SeasonNamingConvention string                      `yaml:"SeasonNamingConvention"` // Season naming convention for the media server. Only needed for Plex. Will default to 2
	UserID                 string                      `yaml:"UserID"`                 // User ID for accessing the media server. This is used for Emby and Jellyfin servers.
}

// Config_TMDB represents the configuration for TMDB (The Movie Database) integration.
type Config_TMDB struct {
	ApiKey string `yaml:"ApiKey"` // API key for accessing TMDB services.
}

// Config_Mediux represents the configuration for Mediux integration.
type Config_Mediux struct {
	Token           string `yaml:"Token"`           // Authentication token for accessing Mediux services.
	DownloadQuality string `yaml:"DownloadQuality"` // Quality of the media to download from Mediux (Options: "original", "optimized") Defaults to "optimized".
}

// Config_AutoDownload represents the configuration for auto-download settings.
type Config_AutoDownload struct {
	Enabled bool   `yaml:"Enabled"` // Whether auto-download is enabled.
	Cron    string `yaml:"Cron"`    // Cron expression for scheduling auto-downloads.
}

type Config_Kometa struct {
	RemoveLabels bool     `yaml:"RemoveLabels"` // Whether to remove overlays from images.
	Labels       []string `yaml:"Labels"`       // List of labels to remove from images.
}

type Config_Notifications struct {
	Enabled   bool                            `yaml:"Enabled"`             // Whether this notification method is enabled
	Providers []Config_Notification_Providers `yaml:"Providers,omitempty"` // List of notification providers
}

type Config_Notification_Providers struct {
	Provider string                        `yaml:"Provider"`           // Notification provider
	Enabled  bool                          `yaml:"Enabled"`            // Whether this notification method is enabled
	Discord  *Config_Notification_Discord  `yaml:"Discord,omitempty"`  // Discord notification settings
	Pushover *Config_Notification_Pushover `yaml:"Pushover,omitempty"` // Pushover notification settings
}

type Config_Notification_Discord struct {
	Webhook string `yaml:"Webhook"` // Webhook URL for the Discord notification provider.
}

type Config_Notification_Pushover struct {
	Token   string `yaml:"Token"`   // Token for the Pushover notification provider.
	UserKey string `yaml:"UserKey"` // UserKey for the Pushover notification provider.
}
