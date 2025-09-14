import { ReturnErrorMessage } from "@/services/api.shared";
import apiClient from "@/services/apiClient";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";
import { AppConfig } from "@/types/config";

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
			SaveImageNextToContent: { Enabled: false },
		},
		TMDB: {
			ApiKey: "",
		},
		Kometa: {
			RemoveLabels: false,
			Labels: [],
		},
		Notifications: {
			Enabled: false,
			Providers: [],
		},
	}) satisfies AppConfig;

export const fetchConfig = async (): Promise<APIResponse<AppConfig>> => {
	log("api.settings - Fetching app configuration started");
	try {
		const response = await apiClient.get<APIResponse<AppConfig>>(`/config`);
		log("api.settings - Fetching app configuration succeeded");
		return {
			...response.data,
			data: response.data.data ?? defaultAppConfig(),
		};
	} catch (error) {
		log(
			`api.settings - Fetching app configuration failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		const err = ReturnErrorMessage<AppConfig>(error);
		// Preserve error status + message, but ensure data shape
		return {
			...err,
			data: defaultAppConfig(),
		};
	}
};
