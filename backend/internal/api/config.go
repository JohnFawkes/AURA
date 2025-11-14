package api

// ---   Config Global Variables --- //
// Config is a pointer to the global configuration instance.
// It is used throughout the application to access configuration settings.
var Global_Config *Config = &Config{}
var Global_Config_Loaded bool = false          // By default, the configuration is not loaded until it is loaded and cached.
var Global_Config_Valid bool = false           // By default, the configuration is invalid until it is validated.
var Global_Config_MediuxValid bool = true      // By default, the mediux config is valid until it is validated.
var Global_Config_MediaServerValid bool = true // By default, the mediaserver config is valid until it is validated.

func Config_NewConfig() *Config {
	return &Config{}
}

func init() {
	Global_Config = Config_NewConfig()
}

// Config represents the main application configuration settings.
// It includes options for server setup, caching, data storage, logging,
// and integration with external services such as Plex, TMDB, and MediUX.
type Config struct {
	Auth          Config_Auth              `yaml:"Auth,omitempty"`          // Authentication settings.
	Logging       Config_Logging           `yaml:"Logging,omitempty"`       // Logging configuration settings.
	MediaServer   Config_MediaServer       `yaml:"MediaServer,omitempty"`   // Media server integration settings.
	Mediux        Config_Mediux            `yaml:"Mediux,omitempty"`        // MediUX integration settings.
	AutoDownload  Config_AutoDownload      `yaml:"AutoDownload,omitempty"`  // Auto-download settings.
	Images        Config_Images            `yaml:"Images,omitempty"`        // Image settings.
	TMDB          Config_TMDB              `yaml:"TMDB,omitempty"`          // TMDB (The Movie Database) integration settings.
	LabelsAndTags Config_LabelsAndTags     `yaml:"LabelsAndTags,omitempty"` // Labels and tags settings.
	Notifications Config_Notifications     `yaml:"Notifications,omitempty"` // Notification settings.
	SonarrRadarr  Config_SonarrRadarr_Apps `yaml:"SonarrRadarr,omitempty"`  // List of Sonarr/Radarr instances to integrate with.
}

type Config_Dev struct {
	Enabled   bool   `yaml:"Enabled,omitempty"`   // Whether to enable development mode.
	LocalPath string `yaml:"LocalPath,omitempty"` // Local path for development mode.
}
type Config_Auth struct {
	Enabled  bool   `yaml:"Enabled"`            // Whether to enable authentication.
	Password string `yaml:"Password,omitempty"` // Password for authentication.
}

type Config_Logging struct {
	Level string `yaml:"Level"`          // Logging level (e.g., TRACE, DEBUG, INFO, WARN, ERROR).
	File  string `yaml:"File,omitempty"` // File path for logging output.
}

type Config_MediaServer struct {
	Type      string                      `yaml:"Type"`                // Type of media server (e.g., plex, emby, jellyfin).
	URL       string                      `yaml:"URL"`                 // Base URL of the media server. This is either the IP:Port or the domain name (e.g., plex.domain.com).
	Token     string                      `yaml:"Token"`               // Authentication token for accessing the media server.
	Libraries []Config_MediaServerLibrary `yaml:"Libraries,omitempty"` // List of media server libraries to manage.
	UserID    string                      `yaml:"UserID,omitempty"`    // User ID for accessing the media server. This is used for Emby and Jellyfin servers.
}
type Config_MediaServerLibrary struct {
	Name      string `yaml:"Name,omitempty"`      // Name of the library.
	SectionID string `yaml:"SectionID,omitempty"` // Unique identifier for the library section.
	Type      string `yaml:"Type,omitempty"`      // Type of the library (e.g., movie, show). All other types are ignored.
	Path      string `yaml:"Path,omitempty"`      // Path of the library on the media server.
}

type Config_Mediux struct {
	Token           string `yaml:"Token"`           // Authentication token for accessing MediUX services.
	DownloadQuality string `yaml:"DownloadQuality"` // Quality of the media to download from MediUX (Options: "original", "optimized") Defaults to "optimized".
}

type Config_AutoDownload struct {
	Enabled bool   `yaml:"Enabled"`        // Whether auto-download is enabled.
	Cron    string `yaml:"Cron,omitempty"` // Cron expression for scheduling auto-downloads.
}

type Config_Images struct {
	CacheImages       Config_CacheImages       `yaml:"CacheImages"`       // Settings for caching images.
	SaveImagesLocally Config_SaveImagesLocally `yaml:"SaveImagesLocally"` // Settings for saving images locally alongside content.
}

type Config_CacheImages struct {
	Enabled bool `yaml:"Enabled"` // Whether to enable caching of images.
}

type Config_SaveImagesLocally struct {
	Enabled                 bool   `yaml:"Enabled"`                           // Whether to save images next to their content.
	Path                    string `yaml:"Path,omitempty"`                    // By default, this is set to alongside the content. If set, this will override that behavior and save all images to this path.
	SeasonNamingConvention  string `yaml:"SeasonNamingConvention,omitempty"`  // Season naming convention for the media server. Only needed for Plex. Will default to 2
	EpisodeNamingConvention string `yaml:"EpisodeNamingConvention,omitempty"` // Episode naming convention for the media server. Only needed for Plex. Will default to match
}

type Config_TMDB struct {
	APIKey string `yaml:"APIKey,omitempty" json:"-"` // API key for accessing TMDB services.
}

type Config_LabelsAndTags struct {
	Applications []Config_LabelsAndTagsProvider `yaml:"Applications,omitempty"`
}

type Config_LabelsAndTagsProvider struct {
	Application               string   `yaml:"Application,omitempty"`
	Enabled                   bool     `yaml:"Enabled,omitempty"`
	Add                       []string `yaml:"Add,omitempty"`
	Remove                    []string `yaml:"Remove,omitempty"`
	AddLabelsForSelectedTypes bool     `yaml:"AddLabelsForSelectedTypes,omitempty"`
}

type Config_Notifications struct {
	Enabled   bool                           `yaml:"Enabled"`             // Whether this notification method is enabled
	Providers []Config_Notification_Provider `yaml:"Providers,omitempty"` // List of notification providers
}

type Config_Notification_Provider struct {
	Provider string                        `yaml:"Provider,omitempty"` // Notification provider
	Enabled  bool                          `yaml:"Enabled,omitempty"`  // Whether this notification method is enabled
	Discord  *Config_Notification_Discord  `yaml:"Discord,omitempty"`  // Discord notification settings
	Pushover *Config_Notification_Pushover `yaml:"Pushover,omitempty"` // Pushover notification settings
	Gotify   *Config_Notification_Gotify   `yaml:"Gotify,omitempty"`   // Gotify notification settings
	Webhook  *Config_Notification_Webhook  `yaml:"Webhook,omitempty"`  // Webhook notification settings
}

type Config_Notification_Discord struct {
	Webhook string `yaml:"Webhook,omitempty"` // Webhook URL for the Discord notification provider.
}

type Config_Notification_Pushover struct {
	Token   string `yaml:"Token,omitempty"`   // Token for the Pushover notification provider.
	UserKey string `yaml:"UserKey,omitempty"` // UserKey for the Pushover notification provider.
}

type Config_Notification_Gotify struct {
	URL   string `yaml:"URL,omitempty"`   // URL for the Gotify notification provider.
	Token string `yaml:"Token,omitempty"` // Token for the Gotify notification provider.
}

type Config_Notification_Webhook struct {
	URL     string            `yaml:"URL,omitempty"`     // URL for the Webhook notification provider.
	Headers map[string]string `yaml:"Headers,omitempty"` // Headers for the Webhook notification provider.
}

type Config_SonarrRadarr_Apps struct {
	Applications []Config_SonarrRadarrApp `yaml:"Applications,omitempty"` // List of Sonarr/Radarr applications to integrate with.
}

type Config_SonarrRadarrApp struct {
	Type    string `yaml:"Type,omitempty"`    // Type of service (either "sonarr" or "radarr").
	Library string `yaml:"Library,omitempty"` // Name of the Media Server library associated with this Sonarr/Radarr instance.
	URL     string `yaml:"URL,omitempty"`     // Base URL of the Sonarr/Radarr server.
	APIKey  string `yaml:"APIKey,omitempty"`  // API key for accessing the Sonarr/Radarr server.
}
