import { AppConfig } from "@/types/config/config-app";

// Central default (extend with all real sections)
export const defaultAppConfig = (): AppConfig =>
	({
		Auth: {
			Enabled: false,
			Password: "",
		},
		Logging: {
			Level: "",
			File: "",
		},
		MediaServer: {
			Type: "",
			URL: "",
			Token: "",
			Libraries: [],
			UserID: "",
		},
		Mediux: {
			Token: "",
			DownloadQuality: "",
		},
		AutoDownload: {
			Enabled: false,
			Cron: "",
		},
		Images: {
			CacheImages: { Enabled: false },
			SaveImageLocally: { Enabled: false, Path: "" },
		},
		TMDB: {
			ApiKey: "",
		},
		LabelsAndTags: {
			Applications: [],
		},
		Notifications: {
			Enabled: false,
			Providers: [],
		},
	}) satisfies AppConfig;
