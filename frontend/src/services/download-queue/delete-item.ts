import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const deleteFromQueue = async (dbItem: DBMediaItemWithPosterSets): Promise<APIResponse<string>> => {
	log(
		"INFO",
		"API - Media Server",
		"Delete from Queue",
		`Deleting '${dbItem.MediaItem.Title}' (TMDB ID: ${dbItem.MediaItem.TMDB_ID}) from the download queue`
	);
	try {
		const response = await apiClient.delete<APIResponse<string>>(`/download-queue/delete`, { data: dbItem });
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error while deleting from download queue");
		} else {
			log(
				"INFO",
				"API - Media Server",
				"Delete from Queue",
				`Deleted '${dbItem.MediaItem.Title}' (TMDB ID: ${dbItem.MediaItem.TMDB_ID}) from the download queue`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Delete from Queue",
			`Failed to delete '${dbItem.MediaItem.Title}' (TMDB ID: ${dbItem.MediaItem.TMDB_ID}) from the download queue: ${
				dbItem.MediaItem.RatingKey
			}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
};
