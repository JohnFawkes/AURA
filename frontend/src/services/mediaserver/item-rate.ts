import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export const rateMediaItem = async (mediaItem: MediaItem, userRating: number): Promise<APIResponse<string>> => {
	log(
		"INFO",
		"API - Media Server",
		"Rate Item",
		`Rating media item '${mediaItem.title}' (TMDB ID: ${mediaItem.tmdb_id})`
	);
	try {
		const response = await apiClient.patch<APIResponse<string>>(`/mediaserver/rate`, null, {
			params: {
				rating_key: mediaItem.rating_key,
				rating: userRating,
			},
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error refreshing media server item");
		} else {
			log(
				"INFO",
				"API - Media Server",
				"Rate Item",
				`Rated media item '${mediaItem.title}' (TMDB ID: ${mediaItem.tmdb_id})`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Rate Item",
			`Failed to rate media item '${mediaItem.title}' (TMDB ID: ${mediaItem.tmdb_id}): ${
				mediaItem.rating_key
			}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
};
