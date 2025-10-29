import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppOnboardingStatus } from "@/types/config/onboarding";

export const fetchOnboardingStatus = async (): Promise<APIResponse<AppOnboardingStatus>> => {
	try {
		const response = await apiClient.get<APIResponse<AppOnboardingStatus>>(`/config/status`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error fetching onboarding status");
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Onboarding",
			`Failed to fetch onboarding status: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<AppOnboardingStatus>(error);
	}
};
