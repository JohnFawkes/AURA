import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export const postClearTempImagesFolder = async (): Promise<APIResponse<void>> => {
	log("INFO", "API - Settings", "Clear Temp Images", "Clearing temp images folder");
	try {
		const response = await apiClient.post<APIResponse<void>>(`/temp-images/clear`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error clearing temp images folder");
		} else {
			log(
				"INFO",
				"API - Settings",
				"Clear Temp Images",
				"Cleared temp images folder successfully",
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Clear Temp Images",
			`Failed to clear temp images folder: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<void>(error);
	}
};
