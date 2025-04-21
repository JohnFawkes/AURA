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
