import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const postAddToQueue = async (
	dbItem: DBMediaItemWithPosterSets
): Promise<APIResponse<DBMediaItemWithPosterSets>> => {
	log(
		"INFO",
		"API - Media Server",
		"Add to Queue",
		`Adding '${dbItem.MediaItem.Title}' (TMDB ID: ${dbItem.MediaItem.TMDB_ID}) to the download queue`
	);
	try {
		const response = await apiClient.post<APIResponse<DBMediaItemWithPosterSets>>(`/download-queue/add`, dbItem);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error while adding to download queue");
		} else {
			log(
				"INFO",
				"API - Media Server",
				"Add to Queue",
				`Added '${dbItem.MediaItem.Title}' (TMDB ID: ${dbItem.MediaItem.TMDB_ID}) to the download queue`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Add to Queue",
			`Failed to add '${dbItem.MediaItem.Title}' (TMDB ID: ${dbItem.MediaItem.TMDB_ID}) to the download queue: ${
				dbItem.MediaItem.RatingKey
			}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<DBMediaItemWithPosterSets>(error);
	}
};
