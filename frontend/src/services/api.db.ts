import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { ClientMessage } from "../types/clientMessage";

export const fetchAllItemsFromDB = async (): Promise<
	APIResponse<ClientMessage[]>
> => {
	try {
		const response = await apiClient.get<APIResponse<ClientMessage[]>>(
			`/db/get/all`
		);
		return response.data;
	} catch (error: any) {
		// Check if the error has a response from the server
		if (error.response && error.response.data) {
			// Return the server's response
			return error.response.data;
		}

		// If no server response, return a default error response
		return {
			status: "error",
			message: "Failed to fetch info from DB API",
		};
	}
};
