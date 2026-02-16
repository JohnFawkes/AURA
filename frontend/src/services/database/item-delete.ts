import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";
import { DBSavedItem } from "@/types/database/db-poster-set";

export const deleteItemFromDB = async (deleteItem: DBSavedItem): Promise<APIResponse<DBSavedItem>> => {
	log(
		"INFO",
		"API - DB",
		"Update",
		`Patching ${deleteItem.media_item.title} (${deleteItem.media_item.tmdb_id} | ${deleteItem.media_item.library_title}) in DB`,
		deleteItem
	);
	try {
		const response = await apiClient.delete<APIResponse<DBSavedItem>>(`/db`, {
			params: {
				tmdb_id: deleteItem.media_item.tmdb_id,
				library_title: deleteItem.media_item.library_title,
			},
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error patching item in DB");
		} else {
			log(
				"INFO",
				"API - DB",
				"Update",
				`Patched ${deleteItem.media_item.title} (${deleteItem.media_item.tmdb_id} | ${deleteItem.media_item.library_title}) in DB`,
				response.data
			);
		}
		const { clearMediaItemSavedSets } = useLibrarySectionsStore.getState();
		clearMediaItemSavedSets({
			tmdbID: deleteItem.media_item.tmdb_id,
			libraryTitle: deleteItem.media_item.library_title,
		});
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - DB",
			"Update",
			`Failed to patch ${deleteItem.media_item.title} (${deleteItem.media_item.tmdb_id} | ${deleteItem.media_item.library_title}) in DB: ${
				error instanceof Error ? error.message : "Unknown error"
			}`,
			error
		);
		return ReturnErrorMessage<DBSavedItem>(error);
	}
};
