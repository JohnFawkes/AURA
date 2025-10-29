import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

export const fetchMediuxSets = async (
	tmdbID: string,
	itemType: string,
	librarySection: string
): Promise<APIResponse<PosterSet[]>> => {
	log(
		"INFO",
		"API - Mediux",
		"Fetch Sets",
		`Fetching Mediux sets for tmdbID: ${tmdbID}, itemType: ${itemType}, librarySection: ${librarySection}`
	);
	try {
		const response = await apiClient.get<APIResponse<PosterSet[]>>(`/mediux/sets`, {
			params: {
				itemType: itemType,
				librarySection: librarySection,
				tmdbID: tmdbID,
			},
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || `Unknown error fetching Mediux sets for tmdbID: ${tmdbID}`);
		} else {
			log(
				"INFO",
				"API - Mediux",
				"Fetch Sets",
				`Fetched Mediux sets for tmdbID: ${tmdbID} successfully`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Mediux",
			"Fetch Sets",
			`Failed to fetch Mediux sets for tmdbID: ${tmdbID}, itemType: ${itemType}, librarySection: ${librarySection}: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<PosterSet[]>(error);
	}
};
