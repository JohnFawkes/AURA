import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfig } from "@/types/config/config-app";
import { defaultAppConfig } from "@/types/config/config-default-app";

export const fetchConfig = async (): Promise<APIResponse<AppConfig>> => {
	log("INFO", "API - Settings", "Fetch Config", "Fetching app configuration");
	try {
		const response = await apiClient.get<APIResponse<AppConfig>>(`/config`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error fetching app configuration");
		} else {
			log("INFO", "API - Settings", "Fetch Config", "Fetched app configuration successfully", response.data);
		}
		return {
			...response.data,
			data: response.data.data ?? defaultAppConfig(),
		};
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Fetch Config",
			`Failed to fetch app configuration: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		const err = ReturnErrorMessage<AppConfig>(error);
		// Preserve error status + message, but ensure data shape
		return {
			...err,
			data: defaultAppConfig(),
		};
	}
};
