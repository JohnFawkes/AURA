import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppOnboardingStatus } from "@/types/config/onboarding";

export const fetchOnboardingStatus = async (): Promise<APIResponse<AppOnboardingStatus>> => {
	log("api.settings - Fetching onboarding status started");
	try {
		const response = await apiClient.get<APIResponse<AppOnboardingStatus>>(`/onboarding/status`);
		log("api.settings - Fetching onboarding status succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Fetching onboarding status failed: ${error instanceof Error ? error.message : "Unknown error"}`
		);
		return ReturnErrorMessage<AppOnboardingStatus>(error);
	}
};
