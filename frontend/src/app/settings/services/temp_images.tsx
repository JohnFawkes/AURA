import { ReturnErrorMessage } from "@/services/api.shared";
import apiClient from "@/services/apiClient";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/apiResponse";

export const postClearTempImagesFolder = async (): Promise<APIResponse<void>> => {
	log("api.settings - Clearing temp images folder started");
	try {
		const response = await apiClient.post<APIResponse<void>>(`/temp-images/clear`);
		log("api.settings - Clearing temp images folder succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.settings - Clearing temp images folder failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<void>(error);
	}
};
