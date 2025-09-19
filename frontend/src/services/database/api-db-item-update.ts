import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const patchSavedItemInDB = async (
	saveItem: DBMediaItemWithPosterSets
): Promise<APIResponse<DBMediaItemWithPosterSets>> => {
	log(`api.db - Patching DBMediaItemWithPosterSets for item with ID ${saveItem.MediaItemID} started.`);
	try {
		const response = await apiClient.patch<APIResponse<DBMediaItemWithPosterSets>>(`/db/update/`, saveItem);
		log(`api.db - Patching DBMediaItemWithPosterSets for item with ID ${saveItem.MediaItemID} succeeded`);
		const { updateMediaItem } = useLibrarySectionsStore.getState();
		updateMediaItem(saveItem.MediaItem.RatingKey, saveItem.MediaItem.LibraryTitle, "update");
		return response.data;
	} catch (error) {
		log(
			`api.db - Patching DBMediaItemWithPosterSets for item with ID ${
				saveItem.MediaItemID
			} failed: ${error instanceof Error ? error.message : "Unknown error"}`
		);
		return ReturnErrorMessage<DBMediaItemWithPosterSets>(error);
	}
};
