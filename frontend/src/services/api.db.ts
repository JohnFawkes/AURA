import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { ReturnErrorMessage } from "./api.shared";
import { log } from "@/lib/logger";
import { SavedSet } from "@/types/databaseSavedSet";

export const fetchAllItemsFromDB = async (): Promise<
	APIResponse<SavedSet[]>
> => {
	log("api.db - Fetching all items from the database started");
	try {
		const response = await apiClient.get<APIResponse<SavedSet[]>>(
			`/db/get/all`
		);
		log("api.db - Fetching all items from the database succeeded");
		return response.data;
	} catch (error) {
		log(
			`api.db - Fetching all items from the database failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<SavedSet[]>(error);
	}
};

export const deleteItemFromDB = async (
	id: string
): Promise<APIResponse<SavedSet>> => {
	log(`api.db - Deleting item with ID ${id} started`);
	try {
		const response = await apiClient.delete<APIResponse<SavedSet>>(
			`/db/delete/${id}`
		);
		log(`api.db - Deleting item with ID ${id} succeeded`);
		return response.data;
	} catch (error) {
		log(
			`api.db - Deleting item with ID ${id} failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<SavedSet>(error);
	}
};

export const patchSavedSetInDB = async (
	savedSet: SavedSet
): Promise<APIResponse<SavedSet>> => {
	log(`api.db - Patching SavedSet for item with ID ${savedSet.ID} started.`);
	try {
		const response = await apiClient.patch<APIResponse<SavedSet>>(
			`/db/update/`,
			savedSet
		);
		log(
			`api.db - Patching SavedSet for item with ID ${savedSet.ID} succeeded`
		);
		return response.data;
	} catch (error) {
		log(
			`api.db - Patching SavedSet for item with ID ${
				savedSet.ID
			} failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<SavedSet>(error);
	}
};
