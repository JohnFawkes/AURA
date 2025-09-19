import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterFile } from "@/types/media-and-posters/poster-sets";

export const patchDownloadPosterFileAndUpdateMediaServer = async (
	posterFile: PosterFile,
	mediaItem: MediaItem,
	fileName: string
): Promise<APIResponse<string>> => {
	log(`api.mediaserver - Downloading ${fileName}`, {
		posterFile: posterFile,
		mediaItem: mediaItem,
	});
	try {
		const response = await apiClient.patch<APIResponse<string>>(`/mediaserver/download/file`, {
			PosterFile: posterFile,
			MediaItem: mediaItem,
		});
		return response.data;
	} catch (error) {
		log(
			`api.mediaserver - Download and update failed: ${error instanceof Error ? error.message : "Unknown error"}`
		);
		return ReturnErrorMessage<string>(error);
	}
};
