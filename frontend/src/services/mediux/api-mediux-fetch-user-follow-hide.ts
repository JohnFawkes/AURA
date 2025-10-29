import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediuxUserFollowHide } from "@/types/mediux/mediux-user-follow-hide";

export const fetchMediuxUserFollowHides = async (): Promise<APIResponse<MediuxUserFollowHide>> => {
	log("INFO", "API - Mediux", "Fetch User Follow/Hides", "Fetching user follow/hide data");
	try {
		const response = await apiClient.get<APIResponse<MediuxUserFollowHide>>(`/mediux//user-follow-hiding`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error fetching user follow/hide data");
		} else {
			log(
				"INFO",
				"API - Mediux",
				"Fetch User Follow/Hides",
				"Fetched user follow/hide data successfully",
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Mediux",
			"Fetch User Follow/Hides",
			`Failed to fetch user follow/hide data: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<MediuxUserFollowHide>(error);
	}
};
