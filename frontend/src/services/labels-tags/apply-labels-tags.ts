import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const postApplyLabelsTagsToDBItem = async (dbItem: DBMediaItemWithPosterSets): Promise<APIResponse<string>> => {
	log(
		"INFO",
		"API - Labels/Tags",
		"Apply Labels/Tags",
		`Applying labels/tags to '${dbItem.MediaItem.Title}' (TMDB ID: ${dbItem.MediaItem.TMDB_ID})`
	);
	try {
		const response = await apiClient.post<APIResponse<string>>(`/labels-tags/apply`, dbItem);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error while applying labels/tags");
		} else {
			log(
				"INFO",
				"API - Labels/Tags",
				"Apply Labels/Tags",
				`Applied labels/tags to '${dbItem.MediaItem.Title}' (TMDB ID: ${dbItem.MediaItem.TMDB_ID})`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Labels/Tags",
			"Apply Labels/Tags",
			`Failed to apply labels/tags to '${dbItem.MediaItem.Title}' (TMDB ID: ${dbItem.MediaItem.TMDB_ID}): ${
				dbItem.MediaItem.RatingKey
			}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
};
