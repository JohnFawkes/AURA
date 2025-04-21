import { AxiosError } from "axios";
import { APIResponse } from "../types/apiResponse";

export const ReturnErrorMessage = <T>(error: unknown): APIResponse<T> => {
	console.error("Returning error:", error);
	if (error instanceof AxiosError) {
		return {
			status: "error",
			message: error.response?.data.message || error.message,
		} as APIResponse<T>;
	} else if (error instanceof Error) {
		return {
			status: "error",
			message: error.message,
		} as APIResponse<T>;
	} else {
		return {
			status: "error",
			message: "An unknown error occurred",
		} as APIResponse<T>;
	}
};
