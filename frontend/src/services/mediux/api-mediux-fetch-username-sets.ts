import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediuxUserAllSetsResponse } from "@/types/mediux/mediux-sets";

export const fetchAllUserSets = async (username: string): Promise<APIResponse<MediuxUserAllSetsResponse>> => {
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
