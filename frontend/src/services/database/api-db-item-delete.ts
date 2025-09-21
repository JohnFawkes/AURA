import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const deleteMediaItemFromDB = async (saveItem: DBMediaItemWithPosterSets): Promise<APIResponse<string>> => {
	log("INFO", "API - DB", "Delete", `Deleting media item with ID ${saveItem.MediaItemID} from DB`, saveItem);
	try {
		const response = await apiClient.delete<APIResponse<string>>(`/db/delete/mediaitem/${saveItem.MediaItemID}`);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error deleting media item from DB");
		} else {
			log(
				"INFO",
				"API - DB",
				"Delete",
				`Deleted media item with ID ${saveItem.MediaItemID} from DB`,
				response.data
			);
		}
		const { updateMediaItem } = useLibrarySectionsStore.getState();
		updateMediaItem(saveItem.MediaItem.RatingKey, saveItem.MediaItem.LibraryTitle, "delete");
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - DB",
			"Delete",
			`Failed to delete media item with ID ${saveItem.MediaItemID} from DB: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
};
