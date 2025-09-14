import { ReturnErrorMessage } from "@/services/api.shared";
import apiClient from "@/services/apiClient";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";

export const postSendTestNotification = async (): Promise<APIResponse<string>> => {
	log("api.settings - Sending test notification started");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/health/status/notification`);
		log("api.settings - Sending test notification succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Sending test notification failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
};
