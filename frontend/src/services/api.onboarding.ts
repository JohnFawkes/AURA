import { log } from "@/lib/logger";

import { AppConfig } from "@/types/config";

import { APIResponse } from "../types/apiResponse";
import { ReturnErrorMessage } from "./api.shared";
import apiClient from "./apiClient";

export interface AppOnboardingStatus {
	configLoaded: boolean;
	configValid: boolean;
	needsSetup: boolean;
	currentSetup: AppConfig;
}

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
