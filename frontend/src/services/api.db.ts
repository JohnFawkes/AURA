import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { ReturnErrorMessage } from "./api.shared";
import { log } from "@/lib/logger";
import {
	DBMediaItemWithPosterSets,
	DBSavedItem,
} from "@/types/databaseSavedSet";
import { CACHE_DB_NAME, CACHE_STORE_NAME } from "@/constants/cache";
import { openDB } from "idb";
import { MediaItem } from "@/types/mediaItem";

const updateMediaItemStore = async (
	ratingKey: string,
	sectionTitle: string
): Promise<void> => {
	const db = await openDB(CACHE_DB_NAME, 1, {
		upgrade(db) {
			if (!db.objectStoreNames.contains(CACHE_STORE_NAME)) {
				db.createObjectStore(CACHE_STORE_NAME);
			}
		},
	});

	// Retrieve the library record
	const librarySection = await db.get(CACHE_STORE_NAME, sectionTitle);
	if (
		!librarySection ||
		!librarySection.data ||
		!librarySection.data.MediaItems
	) {
		throw new Error("Library record not found or invalid structure");
	}

	// Find and update the media item with the matching ratingKey
	const mediaItems: MediaItem[] = librarySection.data.MediaItems;
	const index = mediaItems.findIndex((item) => item.RatingKey === ratingKey);
	if (index === -1) {
		throw new Error("Media item not found in library section");
	}

	mediaItems[index].ExistInDatabase = true;

	// Write the updated record back to the store
	await db.put(CACHE_STORE_NAME, librarySection, sectionTitle);
};

export const postAddItemToDB = async (
	SaveItem: DBSavedItem
): Promise<APIResponse<DBSavedItem>> => {
	log("api.db - Adding item to DB started");
	try {
		const response = await apiClient.post<APIResponse<DBSavedItem>>(
			`/db/add/item`,
			SaveItem
		);
		log("api.db - Adding item to DB succeeded");
		// Call updateMediaItemStore and swallow any errors if it fails.
		updateMediaItemStore(
			SaveItem.MediaItem.RatingKey,
			SaveItem.MediaItem.LibraryTitle
		).catch((e) => {
			log(
				`api.db - Updating media item cache failed: ${
					e instanceof Error ? e.message : "Unknown error"
				}`
			);
		});
		return response.data;
	} catch (error) {
		log(
			`api.db - Adding item to DB failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<DBSavedItem>(error);
	}
};

export const fetchAllItemsFromDB = async (): Promise<
	APIResponse<DBMediaItemWithPosterSets[]>
> => {
	log("api.db - Fetching all items from the database started");
	try {
		const response = await apiClient.get<
			APIResponse<DBMediaItemWithPosterSets[]>
		>(`/db/get/all`);
		log("api.db - Fetching all items from the database succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.db - Fetching all items from the database failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<DBMediaItemWithPosterSets[]>(error);
	}
};

export const deleteMediaItemFromDB = async (
	ratingKey: string
): Promise<APIResponse<string>> => {
	log(`api.db - Deleting media item with ID ${ratingKey} started`);
	try {
		const response = await apiClient.delete<APIResponse<string>>(
			`/db/delete/mediaitem/${ratingKey}`
		);
		log(`api.db - Deleting media item with ID ${ratingKey} succeeded`);
		return response.data;
	} catch (error) {
		log(
			`api.db - Deleting media item with ID ${ratingKey} failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
};

export const patchSavedItemInDB = async (
	saveItem: DBMediaItemWithPosterSets
): Promise<APIResponse<DBMediaItemWithPosterSets>> => {
	log(
		`api.db - Patching DBMediaItemWithPosterSets for item with ID ${saveItem.MediaItemID} started.`
	);
	try {
		const response = await apiClient.patch<
			APIResponse<DBMediaItemWithPosterSets>
		>(`/db/update/`, saveItem);
		log(
			`api.db - Patching DBMediaItemWithPosterSets for item with ID ${saveItem.MediaItemID} succeeded`
		);
		return response.data;
	} catch (error) {
		log(
			`api.db - Patching DBMediaItemWithPosterSets for item with ID ${
				saveItem.MediaItemID
			} failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<DBMediaItemWithPosterSets>(error);
	}
};
