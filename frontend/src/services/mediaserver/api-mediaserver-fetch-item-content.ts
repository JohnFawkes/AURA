import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export const fetchMediaServerItemContent = async (
	ratingKey: string,
	sectionTitle: string
): Promise<APIResponse<MediaItem>> => {
	try {
		// Encode sectionTitle to handle spaces and special characters
		const encodedSectionTitle = encodeURIComponent(sectionTitle);

		const response = await apiClient.get<APIResponse<MediaItem>>(
			`/mediaserver/item/${ratingKey}?sectionTitle=${encodedSectionTitle}`
		);
		return response.data;
	} catch (error) {
		log(
			`API - Media Server - Fetching item content for ratingKey ${ratingKey} failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<MediaItem>(error);
	}
};
