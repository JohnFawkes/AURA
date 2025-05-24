import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { PosterSet } from "../types/posterSets";
import { ReturnErrorMessage } from "./api.shared";
import { log } from "@/lib/logger";
import { MediuxUserFollowHide } from "@/types/mediuxUserFollowsHides";

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

export const fetchMediuxUserFollowHides = async (): Promise<
	APIResponse<MediuxUserFollowHide>
> => {
	log(`api.mediux - Fetching Mediux user follow/hide data started`);
	try {
		const response = await apiClient.get<APIResponse<MediuxUserFollowHide>>(
			`/mediux/user/following_hiding`
		);
		log(`api.mediux - Fetching Mediux user follow/hide data completed`);
		return response.data;
	} catch (error) {
		log(
			`api.mediux - Fetching Mediux user follow/hide data failed with error: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<MediuxUserFollowHide>(error);
	}
};
