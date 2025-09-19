import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

export const fetchMediuxSets = async (
	tmdbID: string,
	itemType: string,
	librarySection: string,
	ratingKey: string
): Promise<APIResponse<PosterSet[]>> => {
	log(
		`api.mediux - Fetching Mediux sets for tmdbID: ${tmdbID}, itemType: ${itemType}, librarySection: ${librarySection} started`
	);
	try {
		const response = await apiClient.get<APIResponse<PosterSet[]>>(
			`/mediux/sets/get/${itemType}/${librarySection}/${ratingKey}/${tmdbID}`
		);
		log(
			`api.mediux - Fetching Mediux sets for tmdbID: ${tmdbID}, itemType: ${itemType}}, librarySection: ${librarySection} completed`
		);
		return response.data;
	} catch (error) {
		log(
			`api.mediux - Fetching Mediux sets for tmdbID: ${tmdbID}, itemType: ${itemType} librarySection: ${librarySection} failed with error: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<PosterSet[]>(error);
	}
};
