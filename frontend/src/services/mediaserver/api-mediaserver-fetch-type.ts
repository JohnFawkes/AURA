import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

/**
 * Fetches the media server type (e.g., Plex, Emby, Jellyfin).
 *
 * Initiates a GET request to the `/api/mediaserver/type/` endpoint and returns the response data.
 *
 * @returns {Promise<APIResponse<LibrarySection[]>>} A promise that resolves to the API response containing an array of library sections.
 */
export const fetchMediaServerType = async (): Promise<APIResponse<{ serverType: string }>> => {
	log("api.mediaserver - Fetching server type started");
	try {
		const response = await apiClient.get<APIResponse<{ serverType: string }>>(`/mediaserver/type`);
		log("api.mediaserver - Fetching server type succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.mediaserver - Fetching server type failed: ${error instanceof Error ? error.message : "Unknown error"}`
		);
		return ReturnErrorMessage<{ serverType: string }>(error);
	}
};
