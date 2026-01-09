import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { AppConfigMediaServer } from "@/types/config/config-app";

export async function postMediaServerLibraryOptions(
	mediaServerInfo: AppConfigMediaServer
): Promise<APIResponse<string>> {
	log("INFO", "API - Settings", "Media Server", "Posting media server info to get library options");
	try {
		const response = await apiClient.post<APIResponse<string>>(`/mediaserver/library-options`, mediaServerInfo);
		if (response.data.status === "error") {
			throw new Error(
				response.data.error?.message || "Unknown error posting media server info to get library options"
			);
		} else {
			log(
				"INFO",
				"API - Settings",
				"Media Server",
				"Posted media server info to get library options successfully",
				response.data
			);
		}
		return response.data;
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Media Server",
			`Failed to post media server info to get library options: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		return ReturnErrorMessage<string>(error);
	}
}

export const fetchMediaServerLibraryOptions = async (
	mediaServerInfo: AppConfigMediaServer
): Promise<{ ok: boolean; data: string[] }> => {
	try {
		const loadingToast = toast.loading(`Fetching library options for ${mediaServerInfo.Type}...`);
		const response = await postMediaServerLibraryOptions(mediaServerInfo);
		if (response.status === "error") {
			toast.dismiss(loadingToast);
			return {
				ok: false,
				data: [],
			};
		}
		toast.dismiss(loadingToast);
		toast.success(`Successfully fetched library options for ${mediaServerInfo.Type}`, { duration: 1000 });
		const data: string[] = Array.isArray(response.data)
			? response.data
			: typeof response.data === "string"
				? [response.data]
				: [];
		return { ok: true, data };
	} catch (error) {
		const errorResponse = ReturnErrorMessage<string>(error);
		toast.error(errorResponse.error?.message || "Couldn't connect to media server. Check the URL and Token", {
			duration: 1000,
		});
		return {
			ok: false,
			data: [],
		};
	}
};

export const fetchPinCodeAndIDFromBackend = async (): Promise<{ ok: boolean; pinCode: string; plexID: string }> => {
	log("INFO", "API - Settings", "Media Server", "Fetching Plex PIN code and ID from backend");
	try {
		const response = await apiClient.get<APIResponse<{ pinCode: string; plexID: string }>>(`/config/plex/get-pin`);
		if (response.data.status === "error") {
			return {
				ok: false,
				pinCode: "",
				plexID: "",
			};
		}
		log(
			"INFO",
			"API - Settings",
			"Media Server",
			"Fetched Plex PIN code and ID from backend successfully",
			response.data
		);
		return { ok: true, pinCode: response.data.data?.pinCode || "", plexID: response.data.data?.plexID || "" };
	} catch (error) {
		log(
			"ERROR",
			"API - Settings",
			"Media Server",
			`Failed to fetch Plex PIN code and ID from backend: ${error instanceof Error ? error.message : "Unknown error"}`,
			error
		);
		const errorResponse = ReturnErrorMessage<{ pinCode: string; plexID: string }>(error);
		toast.error(errorResponse.error?.message || "Couldn't fetch Plex Pin Code", {
			duration: 1000,
		});
		return {
			ok: false,
			pinCode: "",
			plexID: "",
		};
	}
};

export const fetchCheckAuthStatusWithPlex = async (
	plexID: string
): Promise<{ ok: boolean; authenticated: boolean; authToken: string; connectionsAvailable: PlexServersResponse[] }> => {
	log("INFO", "API - Settings", "Media Server", "Checking authentication status with Plex");
	try {
		const response = await apiClient.get<
			APIResponse<{ authenticated: boolean; authToken: string; connectionsAvailable: PlexServersResponse[] }>
		>(`/config/plex/check-pin`, {
			params: { plexID },
		});
		if (response.data.status === "error") {
			return {
				ok: false,
				authenticated: false,
				authToken: "",
				connectionsAvailable: [],
			};
		}
		log(
			"INFO",
			"API - Settings",
			"Media Server",
			"Checked authentication status with Plex successfully",
			response.data
		);
		return {
			ok: true,
			authenticated: response.data.data?.authenticated || false,
			authToken: response.data.data?.authToken || "",
			connectionsAvailable: response.data.data?.connectionsAvailable || [],
		};
	} catch {
		log("ERROR", "API - Settings", "Media Server", `Failed to check authentication status with Plex`);
		return {
			ok: false,
			authenticated: false,
			authToken: "",
			connectionsAvailable: [],
		};
	}
};

export interface PlexServersResponse {
	name: string;
	owned: boolean;
	connections: PlexServerConnection[];
}

export interface PlexServerConnection {
	protocol: string;
	address: string;
	port: number;
	uri: string;
	local: boolean;
	relay: boolean;
	ipv6: boolean;
}
