import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export const fetchLogContents = async (): Promise<APIResponse<string>> => {
	log("api.settings - Fetching log contents started");
	try {
		const response = await apiClient.get<APIResponse<string>>(`/logs`);
		log("api.settings - Fetching log contents succeeded");
		return response.data;
	} catch (error) {
		log(`api.settings - Fetching log contents failed: ${error instanceof Error ? error.message : "Unknown error"}`);
		return ReturnErrorMessage<string>(error);
	}
};

export const postClearOldLogs = async (clearToday: boolean = false): Promise<APIResponse<void>> => {
	log("api.settings - Clearing old logs started");
	try {
		const response = await apiClient.post<APIResponse<void>>(`/logs/clear`, { clearToday });
		log("api.settings - Clearing old logs succeeded");
		return response.data;
	} catch (error) {
		log(`api.settings - Clearing old logs failed: ${error instanceof Error ? error.message : "Unknown error"}`);
		return ReturnErrorMessage<void>(error);
	}
};
