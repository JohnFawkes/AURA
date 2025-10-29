import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse, LogErrorInfo } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSet } from "@/types/media-and-posters/poster-sets";
import { MediuxUserFollowHide } from "@/types/mediux/mediux-user-follow-hide";

export const fetchMediaServerItemContent = async (
	ratingKey: string,
	sectionTitle: string,
	returnType: "full" | "mediaitem" = "full"
): Promise<
	APIResponse<{
		serverType: string;
		mediaItem: MediaItem;
		posterSets: PosterSet[];
		userFollowHide: MediuxUserFollowHide;
		error: LogErrorInfo | null;
	}>
> => {
	log(
		"INFO",
		"API - Media Server",
		"Fetch",
		`Fetching item content for ratingKey ${ratingKey} in section ${sectionTitle}`
	);
	try {
		// Encode sectionTitle to handle spaces and special characters
		const params = new URLSearchParams({
			ratingKey: ratingKey,
			sectionTitle: sectionTitle,
			returnType: returnType,
		});
		const response = await apiClient.get<
			APIResponse<{
				serverType: string;
				mediaItem: MediaItem;
				posterSets: PosterSet[];
				userFollowHide: MediuxUserFollowHide;
				error: LogErrorInfo | null;
			}>
		>(`/mediaserver/item`, { params });
		if (response.data.status === "error") {
			throw new Error(
				response.data.error?.message || `Unknown error fetching item content for ratingKey ${ratingKey}`
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
		return ReturnErrorMessage<{
			serverType: string;
			mediaItem: MediaItem;
			posterSets: PosterSet[];
			userFollowHide: MediuxUserFollowHide;
			error: null;
		}>(error);
	}
};
