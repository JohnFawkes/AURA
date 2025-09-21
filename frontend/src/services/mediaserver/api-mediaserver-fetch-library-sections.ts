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
	log("INFO", "API - Media Server", "Fetch Library Sections", "Fetching all library sections");
	try {
		const response = await apiClient.get<APIResponse<LibrarySection[]>>(`/mediaserver/sections/`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error fetching all library sections");
		} else {
			log("INFO", "API - Media Server", "Fetch Library Sections", `Fetched all library sections successfully`);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Fetch Library Sections",
			`Failed to fetch all library sections: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<LibrarySection[]>(error);
	}
};
