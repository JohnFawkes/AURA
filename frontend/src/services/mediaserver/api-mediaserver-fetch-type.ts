import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

/**
 * Fetches the media server type (e.g., Plex, Emby, Jellyfin).
 *
 * Initiates a GET request to the `/config/mediaserver/type` endpoint and returns the response data.
 *
 * @returns {Promise<APIResponse<LibrarySection[]>>} A promise that resolves to the API response containing an array of library sections.
 */
export const fetchMediaServerType = async (): Promise<APIResponse<{ serverType: string }>> => {
	log("INFO", "API - Media Server", "Fetch Server Type", "Fetching media server type");
	try {
		const response = await apiClient.get<APIResponse<{ serverType: string }>>(`/config/mediaserver/type`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error fetching media server type");
		} else {
			log(
				"INFO",
				"API - Media Server",
				"Fetch Server Type",
				`Fetched media server type successfully`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Fetch Server Type",
			`Failed to fetch media server type: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<{ serverType: string }>(error);
	}
};
