import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";
import { useLibrarySectionsStore } from "@/lib/stores/global-store-library-sections";

import { APIResponse } from "@/types/api/api-response";
import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

export const postAddItemToDB = async (
	saveItem: DBMediaItemWithPosterSets
): Promise<APIResponse<DBMediaItemWithPosterSets>> => {
	let largeDataSize = false;
	if (JSON.stringify(saveItem.PosterSets).length / 1024 / 1024 > 10) {
		largeDataSize = true;
		saveItem.PosterSets.forEach((posterSet) => {
			posterSet.PosterSet.Poster = undefined;
			posterSet.PosterSet.OtherPosters = [];
			posterSet.PosterSet.Backdrop = undefined;
			posterSet.PosterSet.OtherBackdrops = [];
			posterSet.PosterSet.SeasonPosters = [];
			posterSet.PosterSet.TitleCards = [];
		});
	}
	log(
		"INFO",
		"API - DB",
		"Add",
		`Adding '${saveItem.MediaItem.Title} (${saveItem.TMDB_ID} | ${saveItem.LibraryTitle})' to DB`,
		{
			"Large Data Size": largeDataSize,
		}
	);
	try {
		const response = await apiClient.post<APIResponse<DBMediaItemWithPosterSets>>(`/db/add`, {
			saveItem,
			largeDataSize,
		});
		if (response.data.status === "error") {
			throw new Error(response.data.error?.message || "Unknown error adding item to DB");
		} else {
			log(
				"INFO",
				"API - DB",
				"Add",
				`Added '${saveItem.MediaItem.Title} (${saveItem.TMDB_ID} | ${saveItem.LibraryTitle})' to DB successfully`,
				response.data
			);
		}

		const { updateMediaItem } = useLibrarySectionsStore.getState();
		if (response.data.data?.MediaItem) {
			updateMediaItem(response.data.data.MediaItem, "add");
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
		return ReturnErrorMessage<DBMediaItemWithPosterSets>(error);
	}
};
