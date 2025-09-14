import { ReturnErrorMessage } from "@/services/api.shared";
import apiClient from "@/services/apiClient";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";
import { AppConfig } from "@/types/config";

export const finalizeOnboarding = async (newConfig: AppConfig): Promise<APIResponse<AppConfig>> => {
	log("api.settings - Finalizing app configuration started");
	try {
		const response = await apiClient.post<APIResponse<AppConfig>>(`/onboarding/complete`, newConfig);
		log("api.settings - Finalizing app configuration succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Finalizing app configuration failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<AppConfig>(error);
	}
};
