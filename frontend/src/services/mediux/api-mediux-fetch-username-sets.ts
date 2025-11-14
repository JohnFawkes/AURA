import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediuxUserAllSetsResponse } from "@/types/mediux/mediux-sets";

export const fetchAllUserSets = async (username: string): Promise<APIResponse<MediuxUserAllSetsResponse>> => {
	log("INFO", "API - MediUX", "Fetch All User Sets", `Fetching all user sets for ${username}`);
	try {
		const response = await apiClient.get<APIResponse<MediuxUserAllSetsResponse>>("/mediux/sets-by-user", {
			params: {
				username: username,
			},
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || `Unknown error fetching all user sets for ${username}`);
		} else {
			log(
				"INFO",
				"API - MediUX",
				"Fetch All User Sets",
				`Fetched all user sets for ${username} successfully`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - MediUX",
			"Fetch All User Sets",
			`Failed to fetch all user sets for ${username}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<MediuxUserAllSetsResponse>(error);
	}
};
