import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

export const fetchSetByID = async (
	librarySection: string,
	itemRatingKey: string,
	setID: string,
	itemType: "movie" | "show" | "collection"
): Promise<APIResponse<PosterSet>> => {
	log(`api.mediux - Fetching set by ID: ${setID} started`);
	try {
		const response = await apiClient.get<APIResponse<PosterSet>>(`/mediux/sets/get_set/${setID}`, {
			params: {
				itemType: itemType,
				librarySection: librarySection,
				itemRatingKey: itemRatingKey,
			},
		});
		log(`api.mediux - Fetching set by ID: ${setID} completed`);
		return response.data;
	} catch (error) {
		log(
			`api.mediux - Fetching set by ID: ${setID} failed with error: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<PosterSet>(error);
	}
};
