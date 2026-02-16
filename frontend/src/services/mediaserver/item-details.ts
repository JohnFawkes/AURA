import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSetsResponse } from "@/types/media-and-posters/sets";
import { MediuxUserInfo } from "@/types/mediux/mediux-user-follow-hide";

interface getItemContentResponse {
	server_type: string;
	media_item: MediaItem;
	poster_sets: PosterSetsResponse;
	user_follow_hide: MediuxUserInfo[];
}

export const getMediaItemContent = async (
	itemTitle: string,
	ratingKey: string,
	libraryTitle: string,
	returnType: "full" | "item" = "full"
): Promise<APIResponse<getItemContentResponse>> => {
	log(
		"INFO",
		"API - Media Server",
		"Fetch",
		`Fetching ${returnType} content for '${itemTitle}' [${ratingKey}] from library '${libraryTitle}'`
	);
	try {
		const response = await apiClient.get<APIResponse<getItemContentResponse>>(`/mediaserver/item`, {
			params: {
				rating_key: ratingKey,
				return_type: returnType,
			},
		});
		if (response.data.status === "error" && response.data.data?.media_item == null) {
			throw new Error(
				response.data.error?.message || `Unknown error fetching item content for ratingKey ${ratingKey}`
			);
		} else {
			log(
				"INFO",
				"API - Media Server",
				"Fetch",
				`Fetched ${returnType} content for '${itemTitle}' [${ratingKey}] from library '${libraryTitle}'`
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Fetch",
			`Failed to fetch ${returnType} content for '${itemTitle}' [${ratingKey}]: ${
				error instanceof Error ? error.message : "Unknown error"
			}`,
			error
		);
		return ReturnErrorMessage<getItemContentResponse>(error);
	}
};
