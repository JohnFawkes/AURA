import apiClient from "./apiClient";
import { ClientMessage } from "../types/clientMessage";
import { APIResponse } from "../types/apiResponse";
import { ReturnErrorMessage } from "./api.shared";
import { log } from "../lib/logger"; // Assuming you have a logger utility

export const fetchAllItemsFromDB = async (): Promise<
	APIResponse<ClientMessage[]>
> => {
	log("api.db - Fetching all items from the database started");
	try {
		const response = await apiClient.get<APIResponse<ClientMessage[]>>(
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
		return ReturnErrorMessage<ClientMessage[]>(error);
	}
};

export const deleteItemFromDB = async (
	id: string
): Promise<APIResponse<ClientMessage>> => {
	log(`api.db - Deleting item with ID ${id} started`);
	try {
		const response = await apiClient.delete<APIResponse<ClientMessage>>(
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
		return ReturnErrorMessage<ClientMessage>(error);
	}
};

export const patchSelectedTypesInDB = async (
	id: string,
	selectedTypes: string[]
): Promise<APIResponse<ClientMessage>> => {
	log(
		`api.db - Patching selected types for item with ID ${id} started. Selected types: ${JSON.stringify(
			selectedTypes
		)}`
	);
	try {
		const response = await apiClient.patch<APIResponse<ClientMessage>>(
			`/db/update/${id}`,
			{ selectedTypes }
		);
		log(
			`api.db - Patching selected types for item with ID ${id} succeeded`
		);
		return response.data;
	} catch (error) {
		log(
			`api.db - Patching selected types for item with ID ${id} failed: ${
				error instanceof Error ? error.message : "Unknown error"
			}`
		);
		return ReturnErrorMessage<ClientMessage>(error);
	}
};
