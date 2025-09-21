import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export const fetchMediaServerItemContent = async (
	ratingKey: string,
	sectionTitle: string
): Promise<APIResponse<MediaItem>> => {
	log(
		"INFO",
		"API - Media Server",
		"Fetch",
		`Fetching item content for ratingKey ${ratingKey} in section ${sectionTitle}`
	);
	try {
		// Encode sectionTitle to handle spaces and special characters
		const encodedSectionTitle = encodeURIComponent(sectionTitle);
		const response = await apiClient.get<APIResponse<MediaItem>>(
			`/mediaserver/item/${ratingKey}?sectionTitle=${encodedSectionTitle}`
		);
		if (response.data.status === "error") {
			throw new Error(
				response.data.error?.Message || `Unknown error fetching item content for ratingKey ${ratingKey}`
			);
		} else {
			log("INFO", "API - Media Server", "Fetch", `Fetched item content for ratingKey ${ratingKey}`);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Fetch",
			`Fetching item content for ratingKey ${ratingKey} failed: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<MediaItem>(error);
	}
};
