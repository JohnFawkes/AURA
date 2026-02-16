import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { ImageFile } from "@/types/media-and-posters/sets";

export const downloadImageFileForMediaItem = async (
	imageFile: ImageFile,
	mediaItem: MediaItem,
	fileName: string
): Promise<APIResponse<string>> => {
	log(
		"INFO",
		"API - Media Server",
		"Download and Update",
		`Downloading file '${fileName}' and updating '${mediaItem.title}' (TMDB ID: ${mediaItem.tmdb_id})`
	);
	try {
		const response = await apiClient.post<APIResponse<string>>(`/download/image/item`, {
			image_file: imageFile,
			media_item: mediaItem,
		});
		if (response.data.status === "error") {
			throw new Error(
				response.data.error?.message || "Unknown error downloading poster file and updating media server"
			);
		} else {
			log(
				"INFO",
				"API - Media Server",
				"Download and Update",
				`Downloaded file '${fileName}' and updated '${mediaItem.title}' (TMDB ID: ${mediaItem.tmdb_id})`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Download and Update",
			`Failed to download file '${fileName}' and update media item '${mediaItem.title}' (TMDB ID: ${mediaItem.tmdb_id}): ${
				mediaItem.rating_key
			}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
};
