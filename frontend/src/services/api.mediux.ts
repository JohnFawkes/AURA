import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { PosterSets } from "../types/posterSets";
import { ReturnErrorMessage } from "./api.shared";
import { log } from "@/lib/logger";

export const fetchMediuxSets = async (
	tmdbID: string,
	itemType: string
): Promise<APIResponse<PosterSets>> => {
	log(
		`api.mediux - Fetching Mediux sets for tmdbID: ${tmdbID}, itemType: ${itemType} started`
	);
	try {
		const response = await apiClient.get<APIResponse<PosterSets>>(
			`/mediux/sets/get/${itemType}/${tmdbID}`
		);
		log(
			`api.mediux - Fetching Mediux sets for tmdbID: ${tmdbID}, itemType: ${itemType} succeeded`
		);
		return response.data;
	} catch (error) {
		log(
			`api.mediux - Fetching Mediux sets for tmdbID: ${tmdbID}, itemType: ${itemType} failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<PosterSets>(error);
	}
};
