import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfig } from "@/types/config/config-app";

export const updateConfig = async (newConfig: AppConfig): Promise<APIResponse<AppConfig>> => {
	log("INFO", "API - Settings", "Update Config", "Updating app configuration");
	try {
		const response = await apiClient.post<APIResponse<AppConfig>>(`/config/update`, newConfig);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error updating app configuration");
		} else {
			log("INFO", "API - Settings", "Update Config", "Updated app configuration successfully", response.data);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Update Config",
			`Failed to update app configuration: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<AppConfig>(error);
	}
};
