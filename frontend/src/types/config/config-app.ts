export interface AppConfig {
	Auth: AppConfigAuth;
	Logging: AppConfigLogging; // Logging configuration settings
	MediaServer: AppConfigMediaServer; // Media server integration settings
	Mediux: AppConfigMediux; // MediUX integration settings
	AutoDownload: AppConfigAutoDownload; // Auto-download settings
	Images: AppConfigImages;
	TMDB: AppConfigTMDB; // TMDB (The Movie Database) integration settings
	LabelsAndTags: AppConfigLabelsAndTags; // Labels and tags management settings
	Notifications: AppConfigNotifications; // Notification settings
	SonarrRadarr: AppConfigSonarrRadarrApps; // List of Sonarr/Radarr instances to integrate with
}

export interface AppConfigAuth {
	Enabled: boolean; // Whether authentication is enabled
	Password: string; // Hashed password for authentication
}

export interface AppConfigLogging {
	Level: string; // Logging level (e.g., DEBUG, INFO, WARN, ERROR)
	File?: string; // Log file path
}

export interface AppConfigMediaServer {
	Type: string; // Type of media server (e.g., plex, emby, jellyfin)
	URL: string; // Base URL of the media server
	Token: string; // Authentication token for accessing the media server
	Libraries: AppConfigMediaServerLibrary[]; // List of media server libraries to manage
	UserID?: string; // User ID for accessing the media server (optional for Emby/Jellyfin)
}
export interface AppConfigMediaServerLibrary {
	Name: string; // Name of the library
	SectionID: string; // Unique identifier for the library section
	Type: string; // Type of the library (e.g., movie, show)
}

export interface AppConfigMediux {
	Token: string; // Authentication token for accessing MediUX services
	DownloadQuality: string; // Preferred download quality (e.g., "original", "optimized")
}

export interface AppConfigAutoDownload {
	Enabled: boolean; // Whether auto-download is enabled
	Cron: string; // Cron expression for scheduling auto-downloads
}

export interface AppConfigImages {
	CacheImages: AppConfigCacheImages;
	SaveImagesLocally: AppConfigSaveImagesLocally;
}

export interface AppConfigCacheImages {
	Enabled: boolean; // Whether to enable caching of images.
}

export interface AppConfigSaveImagesLocally {
	Enabled: boolean; // Whether to save images locally.
	Path: string; // Path to save images locally. If empty, images will be saved next to content.
	SeasonNamingConvention: string; // Naming convention for season images.
	EpisodeNamingConvention: string; // Naming convention for episode images.
}

export interface AppConfigTMDB {
	ApiKey: string; // API key for accessing TMDB services
}

export interface AppConfigLabelsAndTags {
	Applications: AppConfigLabelsAndTagsApplication[];
}

export interface AppConfigLabelsAndTagsApplication {
	Application: string; // Name of the application (e.g., Plex)
	Enabled: boolean; // Whether label/tag management is enabled for this application
	Add: string[]; // List of labels/tags to add
	Remove: string[]; // List of labels/tags to remove
}

export interface AppConfigNotifications {
	Enabled: boolean;
	Providers: AppConfigNotificationProviders[];
}

export interface AppConfigNotificationProviders {
	Provider: string;
	Enabled: boolean;
	Discord?: AppConfigNotificationDiscord;
	Pushover?: AppConfigNotificationPushover;
	Gotify?: AppConfigNotificationGotify;
	Webhook?: AppConfigNotificationWebhook;
}

export interface AppConfigNotificationDiscord {
	Enabled: boolean;
	Webhook: string;
}

export interface AppConfigNotificationPushover {
	Enabled: boolean;
	UserKey: string;
	Token: string;
}

export interface AppConfigNotificationGotify {
	Enabled: boolean;
	URL: string;
	Token: string;
}

export interface AppConfigNotificationWebhook {
	Enabled: boolean;
	URL: string;
	Headers: { [key: string]: string }; // Key-value pairs for custom headers
}

export interface AppConfigSonarrRadarrApps {
	Applications: AppConfigSonarrRadarrApp[];
}
export interface AppConfigSonarrRadarrApp {
	Type: string; // Type of service (either "sonarr" or "radarr").
	Library: string; // Name of the Media Server library associated with this Sonarr/Radarr instance.
	URL: string; // Base URL of the Sonarr/Radarr server.
	APIKey: string; // API key for accessing the Sonarr/Radarr server.
}
