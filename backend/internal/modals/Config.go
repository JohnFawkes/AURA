package modals

// Config represents the main application configuration settings.
// It includes options for server setup, caching, data storage, logging,
// and integration with external services such as Plex, TMDB, and Mediux.
type Config struct {
	Port                   int                 `yaml:"Port"`                   // Port on which the server will run.
	CacheImages            bool                `yaml:"CacheImages"`            // Whether to cache images locally.
	SaveImageNextToContent bool                `yaml:"SaveImageNextToContent"` // Whether to save images next to the associated content.
	Logging                Config_Logging      `yaml:"Logging"`                // Logging configuration settings.
	Plex                   Config_Plex         `yaml:"Plex"`                   // Plex integration settings.
	TMDB                   Config_TMDB         `yaml:"TMDB"`                   // TMDB (The Movie Database) integration settings.
	Mediux                 Config_Mediux       `yaml:"Mediux"`                 // Mediux integration settings.
	AutoDownload           Config_AutoDownload `yaml:"AutoDownload"`           // Auto-download settings.
}

// Config_Logging represents the logging configuration settings.
type Config_Logging struct {
	Level string `yaml:"Level"` // Logging level (e.g., DEBUG, INFO, WARN, ERROR).
}

// Config_PlexLibrary represents a single Plex library configuration.
type Config_PlexLibrary struct {
	Name      string `yaml:"Name"`      // Name of the Plex library.
	SectionID string `yaml:"SectionID"` // Unique identifier for the Plex library section.
	Type      string `yaml:"Type"`      // Type of the library (e.g., movie, show). All other types are ignored.
}

// Config_Plex represents the configuration for Plex integration.
type Config_Plex struct {
	URL       string               `yaml:"URL"`       // Base URL of the Plex server. This is either the IP:Port or the domain name (e.g., plex.domain.com).
	Token     string               `yaml:"Token"`     // Authentication token for accessing the Plex server.
	Libraries []Config_PlexLibrary `yaml:"Libraries"` // List of Plex libraries to manage.
}

// Config_TMDB represents the configuration for TMDB (The Movie Database) integration.
type Config_TMDB struct {
	ApiKey string `yaml:"ApiKey"` // API key for accessing TMDB services.
}

// Config_Mediux represents the configuration for Mediux integration.
type Config_Mediux struct {
	Token string `yaml:"Token"` // Authentication token for accessing Mediux services.
}

// Config_AutoDownload represents the configuration for auto-download settings.
type Config_AutoDownload struct {
	Enabled bool   `yaml:"Enabled"` // Whether auto-download is enabled.
	Cron    string `yaml:"Cron"`    // Cron expression for scheduling auto-downloads.
}
