import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";
import { DBSavedItem } from "@/types/database/db-saved-item";

export const postAddItemToDB = async (saveItem: DBSavedItem): Promise<APIResponse<DBSavedItem>> => {
	log("INFO", "API - DB", "Add", `Adding item with RatingKey ${saveItem.MediaItem.RatingKey} to DB`);
	try {
		const response = await apiClient.post<APIResponse<DBSavedItem>>(`/db/add/item`, saveItem);
		if (response.data.status === "error") {
			throw new Error(response.data.error?.Message || "Unknown error adding item to DB");
		} else {
			log(
				"INFO",
				"API - DB",
				"Add",
				`Added item with RatingKey ${saveItem.MediaItem.RatingKey} to DB`,
				response.data
			);
		}
		const { updateMediaItem } = useLibrarySectionsStore.getState();
		updateMediaItem(saveItem.MediaItem.RatingKey, saveItem.MediaItem.LibraryTitle, "add");
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - DB",
			"Add",
			`Failed to add item to DB: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<DBSavedItem>(error);
	}
};
