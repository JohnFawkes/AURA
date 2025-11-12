import { CollectionItem } from "@/app/collections/page";
import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse, LogErrorInfo } from "@/types/api/api-response";
import { PosterFile } from "@/types/media-and-posters/poster-sets";
import { MediuxUserInfo } from "@/types/mediux/mediux-user-follow-hide";

interface fetchCollectionChildrenAndPostersResponse {
	collection_item: CollectionItem;
	collection_sets: CollectionSet[];
	user_follow_hide: MediuxUserInfo[];
	error: LogErrorInfo | null;
}

export interface CollectionSet {
	ID: string;
	Title: string;
	User: {
		Name: string;
	};
	Posters: PosterFile[];
	Backdrops: PosterFile[];
}

export const fetchCollectionChildrenAndPosters = async (
	collection_item: CollectionItem
): Promise<APIResponse<fetchCollectionChildrenAndPostersResponse>> => {
	log(
		"INFO",
		"API - Media Server",
		"Fetch",
		`Fetching collection children for ${collection_item.Title} (${collection_item.RatingKey}) in section ${collection_item.LibraryTitle}`
	);
	try {
		const response = await apiClient.post<APIResponse<fetchCollectionChildrenAndPostersResponse>>(
			`/mediaserver/collection-children`,
			collection_item
		);
		if (response.data.status === "error") {
			throw new Error(
				response.data.error?.message ||
					`Unknown error fetching collection children for ${collection_item.Title} (${collection_item.RatingKey}) in section ${collection_item.LibraryTitle}`
			);
		} else {
			log(
				"INFO",
				"API - Media Server",
				"Fetch",
				`Fetched collection children for ${collection_item.Title} (${collection_item.RatingKey}) in section ${collection_item.LibraryTitle}`
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Fetch",
			`Fetching item content for ${collection_item.Title} (${collection_item.RatingKey}) in section ${collection_item.LibraryTitle} failed: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<fetchCollectionChildrenAndPostersResponse>(error);
	}
};
