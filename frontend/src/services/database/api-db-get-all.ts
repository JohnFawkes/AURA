import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const fetchAllItemsFromDB = async (): Promise<APIResponse<DBMediaItemWithPosterSets[]>> => {
	log("api.db - Fetching all items from the database started");
	try {
		const response = await apiClient.get<APIResponse<DBMediaItemWithPosterSets[]>>(`/db/get/all`);
		log("api.db - Fetching all items from the database succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.db - Fetching all items from the database failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<DBMediaItemWithPosterSets[]>(error);
	}
};
