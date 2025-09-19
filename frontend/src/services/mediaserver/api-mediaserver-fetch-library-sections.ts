import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { LibrarySection } from "@/types/media-and-posters/media-item-and-library";

/**
 * Fetches all library sections from the media server.
 *
 * Initiates a GET request to the `/api/mediaserver/sections/` endpoint and returns the response data.
 *
 * @returns {Promise<APIResponse<LibrarySection[]>>} A promise that resolves to the API response containing an array of library sections.
 */
export const fetchMediaServerLibrarySections = async (): Promise<APIResponse<LibrarySection[]>> => {
	log("api.mediaserver - Fetching all library sections started");
	try {
		const response = await apiClient.get<APIResponse<LibrarySection[]>>(`/mediaserver/sections/`);
		log("api.mediaserver - Fetching all library sections succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.mediaserver - Fetching all library sections failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<LibrarySection[]>(error);
	}
};
