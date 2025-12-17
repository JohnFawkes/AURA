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
			error: {
				message:
					error.response?.data.message || error.response?.data.error?.message || "Failed to connect to API",
				help: "Please make sure the backend API is running and accessible.",
				function: "Axios Request",
				line_number: -1,
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
