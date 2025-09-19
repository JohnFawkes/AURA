import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfig } from "@/types/config/config-app";

export const updateConfig = async (newConfig: AppConfig): Promise<APIResponse<AppConfig>> => {
	log("api.settings - Updating app configuration started");
	try {
		const response = await apiClient.post<APIResponse<AppConfig>>(`/config/update`, newConfig);
		log("api.settings - Updating app configuration succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Updating app configuration failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<AppConfig>(error);
	}
};
