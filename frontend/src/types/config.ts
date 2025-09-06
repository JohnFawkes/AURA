export interface AppConfig {
	CacheImages: boolean; // Whether to cache images locally
	SaveImageNextToContent: boolean; // Whether to save images next to the associated content
	Logging: AppConfigLogging; // Logging configuration settings
	MediaServer: AppConfigMediaServer; // Media server integration settings
	TMDB: AppConfigTMDB; // TMDB (The Movie Database) integration settings
	Mediux: AppConfigMediux; // Mediux integration settings
	AutoDownload: AppConfigAutoDownload; // Auto-download settings
	Kometa: AppConfigKometa; // Kometa integration settings
	Notifications: AppConfigNotifications; // Notification settings
}

export interface AppConfigLogging {
	Level: string; // Logging level (e.g., DEBUG, INFO, WARN, ERROR)
	File: string; // Log file path
}

export interface AppConfigMediaServerLibrary {
	Name: string; // Name of the library
	SectionID: string; // Unique identifier for the library section
	Type: string; // Type of the library (e.g., movie, show)
}

export interface AppConfigMediaServer {
	Type: string; // Type of media server (e.g., plex, emby, jellyfin)
	URL: string; // Base URL of the media server
	Token: string; // Authentication token for accessing the media server
	Libraries: AppConfigMediaServerLibrary[]; // List of media server libraries to manage
	UserID?: string; // User ID for accessing the media server (optional for Emby/Jellyfin)
}

export interface AppConfigTMDB {
	ApiKey: string; // API key for accessing TMDB services
}

export interface AppConfigMediux {
	Token: string; // Authentication token for accessing Mediux services
	DownloadQuality: string; // Preferred download quality (e.g., "original", "optimized")
}

export interface AppConfigAutoDownload {
	Enabled: boolean; // Whether auto-download is enabled
	Cron: string; // Cron expression for scheduling auto-downloads
	CronText: string; // Human-readable text for the cron expression
}

export interface AppConfigKometa {
	RemoveLabels: boolean; // Whether to remove labels from media items
	Labels: string[]; // List of labels to apply to media items
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
