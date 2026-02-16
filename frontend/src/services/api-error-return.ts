import { AxiosError } from "axios";

import { APIResponse } from "@/types/api/api-response";

export const ReturnErrorMessage = <T>(error: unknown): APIResponse<T> => {
	const defaultError = {
		message: "",
		help: "",
		detail: {},
		function: "",
		line_number: 0,
	};

	if (error instanceof AxiosError) {
		return {
			status: "error",
			error: {
				message:
					error.response?.data.message || error.response?.data.error?.message || "Failed to connect to API",
				help:
					error.response?.data?.help ||
					error.response?.data.error?.help ||
					"Please make sure the aura API is running and accessible.",
				detail: error.response?.data?.detail || error.response?.data.error?.detail || null,
				function: error.response?.data?.function || error.response?.data.error?.function || "Axios Request",
				line_number: error.response?.data?.line_number || error.response?.data.error?.line_number || -1,
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
