import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfig } from "@/types/config/config-app";

export const finalizeOnboarding = async (newConfig: AppConfig): Promise<APIResponse<AppConfig>> => {
	log("INFO", "API - Settings", "Onboarding", "Finalizing app configuration", newConfig);
	try {
		const response = await apiClient.post<APIResponse<AppConfig>>(`/onboarding/finalize`, newConfig);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error finalizing app configuration");
		} else {
			log("INFO", "API - Settings", "Onboarding", "Finalized app configuration successfully", response.data);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Onboarding",
			`Failed to finalize app configuration: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<AppConfig>(error);
	}
};
