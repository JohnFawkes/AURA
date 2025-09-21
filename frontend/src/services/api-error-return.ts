import { AxiosError } from "axios";

import { APIResponse } from "@/types/api/api-response";

export const ReturnErrorMessage = <T>(error: unknown): APIResponse<T> => {
	const defaultError = {
		Message: "",
		HelpText: "",
		Function: "",
	};

	if (error instanceof AxiosError) {
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

	return {
		status: "error",
		elapsed: "0",
		error: defaultError,
	} as APIResponse<T>;
};
