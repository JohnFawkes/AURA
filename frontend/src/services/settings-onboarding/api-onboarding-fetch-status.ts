import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppOnboardingStatus } from "@/types/config/onboarding";

export const fetchOnboardingStatus = async (): Promise<APIResponse<AppOnboardingStatus>> => {
	try {
		const response = await apiClient.get<APIResponse<AppOnboardingStatus>>(`/onboarding/status`);
		return response.data;
	} catch (error) {
		log(`API - Onboarding - Error: ${error instanceof Error ? error.message : "Unknown error"}`);
		return ReturnErrorMessage<AppOnboardingStatus>(error);
	}
};
