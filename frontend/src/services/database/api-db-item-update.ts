import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const patchSavedItemInDB = async (
	saveItem: DBMediaItemWithPosterSets
): Promise<APIResponse<DBMediaItemWithPosterSets>> => {
	log(
		"INFO",
		"API - DB",
		"Update",
		`Patching ${saveItem.MediaItem.Title} (${saveItem.TMDB_ID} | ${saveItem.LibraryTitle}) in DB`,
		saveItem
	);
	try {
		const response = await apiClient.patch<APIResponse<DBMediaItemWithPosterSets>>(`/db/update/`, saveItem);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error patching item in DB");
		} else {
			log(
				"INFO",
				"API - DB",
				"Update",
				`Patched ${saveItem.MediaItem.Title} (${saveItem.TMDB_ID} | ${saveItem.LibraryTitle}) in DB`,
				response.data
			);
		}
		const { updateMediaItem } = useLibrarySectionsStore.getState();
		updateMediaItem(saveItem.MediaItem, "update");
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - DB",
			"Update",
			`Failed to patch ${saveItem.MediaItem.Title} (${saveItem.TMDB_ID} | ${saveItem.LibraryTitle}) in DB: ${
				error instanceof Error ? error.message : "Unknown error"
			}`,
			error
		);
		return ReturnErrorMessage<DBMediaItemWithPosterSets>(error);
	}
};
