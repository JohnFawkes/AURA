import { log } from "@/lib/logger";
import { librarySectionsStorage } from "@/lib/storage";

import { DBMediaItemWithPosterSets, DBSavedItem } from "@/types/databaseSavedSet";
import { MediaItem } from "@/types/mediaItem";

import { APIResponse } from "../types/apiResponse";
import { ReturnErrorMessage } from "./api.shared";
import apiClient from "./apiClient";

const updateMediaItemStore = async (ratingKey: string, sectionTitle: string, updateType: string): Promise<void> => {
	try {
		// Retrieve the library record from librarySectionsStorage
		const librarySection = await librarySectionsStorage.getItem<{
			data: {
				MediaItems: MediaItem[];
			};
			timestamp: number;
		}>(sectionTitle);

		if (!librarySection || !librarySection.data || !librarySection.data.MediaItems) {
			throw new Error("Library record not found or invalid structure");
		}

		// Find and update the media item with the matching ratingKey
		const mediaItems: MediaItem[] = librarySection.data.MediaItems;
		const index = mediaItems.findIndex((item) => item.RatingKey === ratingKey);
		if (index === -1) {
			throw new Error("Media item not found in library section");
		}

		// Update the ExistInDatabase flag
		if (updateType === "add" || updateType === "update") mediaItems[index].ExistInDatabase = true;
		else if (updateType === "delete") mediaItems[index].ExistInDatabase = false;

		// Write the updated record back to librarySectionsStorage
		await librarySectionsStorage.setItem(sectionTitle, {
			...librarySection,
			data: {
				...librarySection.data,
				MediaItems: mediaItems,
			},
		});

		log(
			`api.db - Updated media item cache successfully (${updateType}) for RatingKey: ${ratingKey} in section: ${sectionTitle}`
		);
	} catch (error) {
		log(`api.db - Error updating media item cache: ${error instanceof Error ? error.message : "Unknown error"}`);
		throw error;
	}
};

export const postAddItemToDB = async (saveItem: DBSavedItem): Promise<APIResponse<DBSavedItem>> => {
	log("api.db - Adding item to DB started");
	try {
		const response = await apiClient.post<APIResponse<DBSavedItem>>(`/db/add/item`, saveItem);
		log("api.db - Adding item to DB succeeded");
		// Call updateMediaItemStore and swallow any errors if it fails.
		updateMediaItemStore(saveItem.MediaItem.RatingKey, saveItem.MediaItem.LibraryTitle, "add").catch((e) => {
			log(`api.db - Updating media item cache failed: ${e instanceof Error ? e.message : "Unknown error"}`);
		});
		return response.data;
	} catch (error) {
		log(`api.db - Adding item to DB failed: ${error instanceof Error ? error.message : "Unknown error"}`);
		return ReturnErrorMessage<DBSavedItem>(error);
	}
};

export const fetchAllItemsFromDB = async (): Promise<APIResponse<DBMediaItemWithPosterSets[]>> => {
	log("api.db - Fetching all items from the database started");
	try {
		const response = await apiClient.get<APIResponse<DBMediaItemWithPosterSets[]>>(`/db/get/all`);
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

export const deleteMediaItemFromDB = async (saveItem: DBMediaItemWithPosterSets): Promise<APIResponse<string>> => {
	log(`api.db - Deleting media item with ID ${saveItem.MediaItemID} started`);
	try {
		const response = await apiClient.delete<APIResponse<string>>(`/db/delete/mediaitem/${saveItem.MediaItemID}`);
		log(`api.db - Deleting media item with ID ${saveItem.MediaItemID} succeeded`);
		// Call updateMediaItemStore and swallow any errors if it fails.
		updateMediaItemStore(saveItem.MediaItem.RatingKey, saveItem.MediaItem.LibraryTitle, "delete").catch((e) => {
			log(`api.db - Updating media item cache failed: ${e instanceof Error ? e.message : "Unknown error"}`);
		});
		return response.data;
	} catch (error) {
		log(
			`api.db - Deleting media item with ID ${saveItem.MediaItemID} failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<string>(error);
	}
};

export const patchSavedItemInDB = async (
	saveItem: DBMediaItemWithPosterSets
): Promise<APIResponse<DBMediaItemWithPosterSets>> => {
	log(`api.db - Patching DBMediaItemWithPosterSets for item with ID ${saveItem.MediaItemID} started.`);
	try {
		const response = await apiClient.patch<APIResponse<DBMediaItemWithPosterSets>>(`/db/update/`, saveItem);
		log(`api.db - Patching DBMediaItemWithPosterSets for item with ID ${saveItem.MediaItemID} succeeded`);
		// Call updateMediaItemStore and swallow any errors if it fails.
		updateMediaItemStore(saveItem.MediaItem.RatingKey, saveItem.MediaItem.LibraryTitle, "update").catch((e) => {
			log(`api.db - Updating media item cache failed: ${e instanceof Error ? e.message : "Unknown error"}`);
		});
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

export interface AutodownloadResult {
	MediaItemTitle: string;
	Sets: AutodownloadSetResult[];
	OverAllResult: "Error" | "Warn" | "Success" | "Skipped";
	OverAllResultMessage: string;
}

export interface AutodownloadSetResult {
	PosterSetID: string;
	Result: "Success" | "Skipped" | "Error";
	Reason: string;
}

export const postForceRecheckDBItemForAutoDownload = async (
	saveItem: DBMediaItemWithPosterSets
): Promise<APIResponse<AutodownloadResult>> => {
	log(`api.db - Forcing recheck for auto-download for item with ID ${saveItem.MediaItemID} started`);
	try {
		const response = await apiClient.post<APIResponse<AutodownloadResult>>(`/db/force/recheck`, {
			Item: saveItem,
		});
		log(`api.db - Forcing recheck for auto-download for item with ID ${saveItem.MediaItemID} succeeded`);
		return response.data;
	} catch (error) {
		log(
			`api.db - Forcing recheck for auto-download for item with ID ${
				saveItem.MediaItemID
			} failed: ${error instanceof Error ? error.message : "Unknown error"}`
		);
		return ReturnErrorMessage<AutodownloadResult>(error);
	}
};
