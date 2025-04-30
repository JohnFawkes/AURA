import { AxiosError } from "axios";
import { APIResponse } from "../types/apiResponse";
import { log } from "@/lib/logger";

export const ReturnErrorMessage = <T>(error: unknown): APIResponse<T> => {
	if (error instanceof AxiosError) {
		log(
			`api.shared - Axios error occurred: ${
				error.response?.data.message || error.message
			}`
		);
		return {
			status: "error",
			message: error.response?.data.message || error.message,
		} as APIResponse<T>;
	} else if (error instanceof Error) {
		log(`api.shared - General error occurred: ${error.message}`);
		return {
			status: "error",
			message: error.message,
		} as APIResponse<T>;
	} else {
		log("api.shared - Unknown error occurred");
		return {
			status: "error",
			message: "An unknown error occurred",
		} as APIResponse<T>;
	}
};
