import { AxiosError } from "axios";

import { APIResponse } from "@/types/api/api-response";

export const ReturnErrorMessage = <T>(error: unknown): APIResponse<T> => {
	const defaultError = {
		message: "",
		help: "",
		function: "",
		line_number: 0,
		detail: {},
	};

	if (error instanceof AxiosError) {
		return {
			status: "error",
			error: error.response?.data.error || {
				message: error.response?.data.message || error.message,
				help: "Please check your connection and try again",
				function: "AxiosRequest",
				line_number: 0,
			},
		} as APIResponse<T>;
	}

	if (error instanceof Error) {
		return {
			status: "error",
			error: {
				message: error.message,
				help: "An unexpected error occurred",
				function: error.stack?.split("\n")[1]?.trim() || "Unknown",
				line_number: 0,
			},
		} as APIResponse<T>;
	}

	if (typeof error === "string") {
		return {
			status: "error",
			error: {
				message: error,
				help: "",
				function: "",
				line_number: 0,
			},
		};
	}

	return {
		status: "error",
		error: defaultError,
	} as APIResponse<T>;
};
