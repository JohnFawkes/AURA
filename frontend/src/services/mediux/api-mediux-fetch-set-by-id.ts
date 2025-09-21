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
	log(
		"INFO",
		"API - Mediux",
		"Fetch Set By ID",
		`Fetching set by ID: ${setID} for itemType: ${itemType}, librarySection: ${librarySection}, itemRatingKey: ${itemRatingKey}`
	);
	try {
		const response = await apiClient.get<APIResponse<PosterSet>>(`/mediux/sets/get_set/${setID}`, {
			params: {
				itemType: itemType,
				librarySection: librarySection,
				itemRatingKey: itemRatingKey,
			},
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || `Unknown error fetching set by ID: ${setID}`);
		} else {
			log("INFO", "API - Mediux", "Fetch Set By ID", `Fetched set by ID: ${setID} successfully`, response.data);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Mediux",
			"Fetch Set By ID",
			`Failed to fetch set by ID: ${setID}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<PosterSet>(error);
	}
};
