import { AxiosError } from "axios";

import { log } from "@/lib/logger";

import { APIResponse } from "../types/apiResponse";

export const ReturnErrorMessage = <T>(error: unknown): APIResponse<T> => {
	const defaultError = {
		Message: "",
		HelpText: "",
		Function: "",
	};

	if (error instanceof AxiosError) {
		log(`api.shared - Axios error occurred: ${error}`);
		return {
			status: "error",
			elapsed: "0",
			error: error.response?.data.error || {
				Message: error.response?.data.message || error.message,
				HelpText: "Please check your connection and try again",
				Function: "AxiosRequest",
				LineNumber: 0,
			},
		} as APIResponse<T>;
	}

	if (error instanceof Error) {
		log(`api.shared - General error occurred: ${error.message}`);
		return {
			status: "error",
			elapsed: "0",
			error: {
				Message: error.message,
				HelpText: "An unexpected error occurred",
				Function: error.stack?.split("\n")[1]?.trim() || "Unknown",
				LineNumber: 0,
			},
		} as APIResponse<T>;
	}

	if (typeof error === "string") {
		log(`api.shared - String error occurred: ${error}`);
		return {
			status: "error",
			elapsed: "0",
			error: {
				Message: error,
				HelpText: "",
				Function: "",
				LineNumber: 0,
			},
		};
	}

	log("api.shared - Unknown error occurred");
	return {
		status: "error",
		elapsed: "0",
		error: defaultError,
	} as APIResponse<T>;
};
