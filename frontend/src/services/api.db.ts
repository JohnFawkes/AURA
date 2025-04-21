import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { ClientMessage } from "../types/clientMessage";
import { AxiosError } from "axios";

export const fetchAllItemsFromDB = async (): Promise<
	APIResponse<ClientMessage[]>
> => {
	try {
		const response = await apiClient.get<APIResponse<ClientMessage[]>>(
			`/db/get/all`
		);
		return response.data;
	} catch (error) {
		console.log("Error fetching data from DB:", error);

		if (error instanceof AxiosError) {
			return {
				status: "error",
				message:
					error.response?.data.message || "An unknown error occurred",
			};
		} else if (error instanceof Error) {
			return {
				status: "error",
				message: error.message,
			};
		} else {
			return {
				status: "error",
				message: "An unknown error occurred",
			};
		}
	}
};
