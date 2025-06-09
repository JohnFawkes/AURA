import { log } from "@/lib/logger";

import { MediuxUserAllSetsResponse } from "@/types/mediuxUserAllSets";
import { MediuxUserFollowHide } from "@/types/mediuxUserFollowsHides";

import { APIResponse } from "../types/apiResponse";
import { PosterSet } from "../types/posterSets";
import { ReturnErrorMessage } from "./api.shared";
import apiClient from "./apiClient";

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

export const fetchMediuxUserFollowHides = async (): Promise<APIResponse<MediuxUserFollowHide>> => {
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

export const fetchAllUserSets = async (
	username: string
): Promise<APIResponse<MediuxUserAllSetsResponse>> => {
	log(`api.mediux - Fetching all user sets for ${username} started`);
	try {
		const response = await apiClient.get<APIResponse<MediuxUserAllSetsResponse>>(
			`/mediux/sets/get_user/sets/${username}`
		);
		log(`api.mediux - Fetching all user sets for ${username} completed`);
		return response.data;
	} catch (error) {
		log(
			`api.mediux - Fetching all user sets for ${username} failed with error: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<MediuxUserAllSetsResponse>(error);
	}
};

export const fetchShowSetByID = async (setID: string): Promise<APIResponse<PosterSet>> => {
	log(`api.mediux - Fetching show set by ID: ${setID} started`);
	try {
		const response = await apiClient.get<APIResponse<PosterSet>>(
			`/mediux/sets/get_set/${setID}`
		);
		log(`api.mediux - Fetching show set by ID: ${setID} completed`);
		return response.data;
	} catch (error) {
		log(
			`api.mediux - Fetching show set by ID: ${setID} failed with error: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<PosterSet>(error);
	}
};
