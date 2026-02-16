import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";
import { DBPosterSetDetail, DBSavedItem } from "@/types/database/db-poster-set";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export const addNewItemToDB = async (
	mediaItem: MediaItem,
	posterSet: DBPosterSetDetail
): Promise<APIResponse<DBSavedItem>> => {
	let complete = true;
	const media_data_size = JSON.stringify(mediaItem).length / 1024 / 1024;
	const poster_set_data_size = JSON.stringify(posterSet).length / 1024 / 1024;
	const data_size = media_data_size + poster_set_data_size;
	if (data_size > 10) {
		posterSet.images = [];
		mediaItem.series = undefined;
		complete = false;
	}
	posterSet.last_downloaded = new Date().toISOString();
	log(
		"INFO",
		"API - DB",
		"Add",
		`Adding '${mediaItem.title} (${mediaItem.tmdb_id} | ${mediaItem.library_title})' to DB`,
		{
			complete: complete,
			data_size_mb: data_size,
		}
	);
	try {
		const response = await apiClient.post<APIResponse<DBSavedItem>>(`/db`, {
			media_item: mediaItem,
			poster_set: posterSet,
			complete: complete,
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error adding item to DB");
		} else {
			log(
				"INFO",
				"API - DB",
				"Add",
				`Added '${mediaItem.title} (${mediaItem.tmdb_id} | ${mediaItem.library_title})' to DB successfully`,
				response.data
			);
		}

		const { upsertMediaItemSavedSet } = useLibrarySectionsStore.getState();
		if (response.data.data?.media_item) {
			upsertMediaItemSavedSet({
				tmdbID: response.data.data.media_item.tmdb_id,
				libraryTitle: response.data.data.media_item.library_title,
				setID: posterSet.id,
				setUser: posterSet.user_created,
				selectedTypes: posterSet.selected_types,
			});
		}
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
