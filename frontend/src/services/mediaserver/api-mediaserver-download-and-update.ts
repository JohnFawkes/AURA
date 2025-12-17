import { CollectionItem } from "@/app/collections/page";
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
	log(
		"INFO",
		"API - Media Server",
		"Download and Update",
		`Downloading file '${fileName}' and updating '${mediaItem.Title}' (TMDB ID: ${mediaItem.TMDB_ID})`
	);
	try {
		const response = await apiClient.patch<APIResponse<string>>(`/mediaserver/download`, {
			PosterFile: posterFile,
			MediaItem: mediaItem,
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
				`Downloaded file '${fileName}' and updated '${mediaItem.Title}' (TMDB ID: ${mediaItem.TMDB_ID})`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Download and Update",
			`Failed to download file '${fileName}' and update media item '${mediaItem.Title}' (TMDB ID: ${mediaItem.TMDB_ID}): ${
				mediaItem.RatingKey
			}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
};

export const patchDownloadCollectionImageAndUpdateMediaServer = async (
	imageType: "poster" | "backdrop",
	collectionItem: CollectionItem,
	posterFile: PosterFile
): Promise<APIResponse<string>> => {
	log(
		"INFO",
		"API - Media Server",
		"Download and Update Collection Image",
		`Downloading ${imageType} image and updating '${collectionItem.Title}' (ID: ${collectionItem.RatingKey})`
	);
	try {
		const response = await apiClient.patch<APIResponse<string>>(`/mediaserver/download-collection`, {
			ImageType: imageType,
			CollectionItem: collectionItem,
			PosterFile: posterFile,
		});
		if (response.data.status === "error") {
			throw new Error(
				response.data.error?.message || `Unknown error downloading ${imageType} image and updating media server`
			);
		} else {
			log(
				"INFO",
				"API - Media Server",
				"Download and Update Collection Image",
				`Downloaded ${imageType} image and updated '${collectionItem.Title}' (ID: ${collectionItem.RatingKey})`,
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Media Server",
			"Download and Update Collection Image",
			`Failed to download ${imageType} image and update collection item '${collectionItem.Title}' (ID: ${collectionItem.RatingKey}): ${
				collectionItem.RatingKey
			}: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
};
