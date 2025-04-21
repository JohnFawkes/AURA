import apiClient from "./apiClient";
import { APIResponse } from "../types/apiResponse";
import { AppConfig } from "../types/config";
import { ReturnErrorMessage } from "./api.shared";

export const fetchConfig = async (): Promise<APIResponse<AppConfig>> => {
	try {
		const response = await apiClient.get<APIResponse<AppConfig>>(`/config`);
		return response.data;
	} catch (error) {
		return ReturnErrorMessage<AppConfig>(error);
	}
};

export const fetchLogContents = async (): Promise<APIResponse<string>> => {
	try {
		const response = await apiClient.get<APIResponse<string>>(`/logs`);
		return response.data;
	} catch (error) {
		return ReturnErrorMessage<string>(error);
	}
};
