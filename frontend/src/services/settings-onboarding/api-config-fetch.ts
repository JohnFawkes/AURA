import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfig } from "@/types/config/config-app";
import { defaultAppConfig } from "@/types/config/config-default-app";

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
