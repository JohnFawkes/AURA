import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export const patchAddRatingToMediaItem = async (
	mediaItem: MediaItem,
	userRating: number
): Promise<APIResponse<string>> => {
	log(
		"INFO",
		"API - Media Server",
		"Rate Item",
		`Rating media item '${mediaItem.Title}' (TMDB ID: ${mediaItem.TMDB_ID})`
	);
	try {
		const response = await apiClient.patch<APIResponse<string>>(`/mediaserver/rate-item`, {
			ratingKey: mediaItem.RatingKey,
			userRating: userRating,
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error refreshing media server item");
		} else {
			log(
				"INFO",
				"API - Media Server",
				"Rate Item",
				`Rated media item '${mediaItem.Title}' (TMDB ID: ${mediaItem.TMDB_ID})`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Rate Item",
			`Failed to rate media item '${mediaItem.Title}' (TMDB ID: ${mediaItem.TMDB_ID}): ${
				mediaItem.RatingKey
			}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
};
