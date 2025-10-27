import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const deleteMediaItemFromDB = async (
	saveItem: DBMediaItemWithPosterSets
): Promise<APIResponse<DBMediaItemWithPosterSets>> => {
	log("INFO", "API - DB", "Delete", `Deleting ${saveItem.MediaItem.Title} from DB`, saveItem);
	try {
		const response = await apiClient.delete<APIResponse<DBMediaItemWithPosterSets>>(
			`/db/delete/mediaitem/${saveItem.MediaItem.TMDB_ID}/${saveItem.MediaItem.LibraryTitle}`
		);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error deleting media item from DB");
		} else {
			log(
				"INFO",
				"API - DB",
				"Delete",
				`Deleted ${saveItem.MediaItem.Title} from DB successfully`,
				response.data
			);
		}
		const { updateMediaItem } = useLibrarySectionsStore.getState();
		updateMediaItem(response.data.data?.MediaItem || saveItem.MediaItem, "delete");
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - DB",
			"Delete",
			`Failed to delete ${saveItem.MediaItem.Title} from DB: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<DBMediaItemWithPosterSets>(error);
	}
};
