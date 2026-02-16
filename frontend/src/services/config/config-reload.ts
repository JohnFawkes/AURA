import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfig } from "@/types/config/config";

export const reloadAppConfig = async (): Promise<APIResponse<AppConfig>> => {
	try {
		const response = await apiClient.patch<APIResponse<AppConfig>>(`/config`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error fetching onboarding status");
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Config",
			"Reload App Config",
			`Failed to reload app config: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<AppConfig>(error);
	}
};
