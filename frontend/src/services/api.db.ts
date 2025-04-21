import apiClient from "./apiClient";
import { ClientMessage } from "../types/clientMessage";
import { APIResponse } from "../types/apiResponse";
import { ReturnErrorMessage } from "./api.shared";

export const fetchAllItemsFromDB = async (): Promise<
	APIResponse<ClientMessage[]>
> => {
	try {
		const response = await apiClient.get<APIResponse<ClientMessage[]>>(
			`/db/get/all`
		);
		return response.data;
	} catch (error) {
		return ReturnErrorMessage<ClientMessage[]>(error);
	}
};

export const deleteItemFromDB = async (
	id: string
): Promise<APIResponse<ClientMessage>> => {
	try {
		const response = await apiClient.delete<APIResponse<ClientMessage>>(
			`/db/delete/${id}`
		);
		return response.data;
	} catch (error) {
		return ReturnErrorMessage<ClientMessage>(error);
	}
};

export const patchSelectedTypesInDB = async (
	id: string,
	selectedTypes: string[]
): Promise<APIResponse<ClientMessage>> => {
	try {
		const response = await apiClient.patch<APIResponse<ClientMessage>>(
			`/db/update/${id}`,
			{ selectedTypes }
		);
		return response.data;
	} catch (error) {
		return ReturnErrorMessage<ClientMessage>(error);
	}
};
