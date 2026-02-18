import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export interface GetPlexPinAndID_Response {
  plex_pin: string;
  plex_id: string;
}

export const GetPlexPinAndID = async (): Promise<APIResponse<GetPlexPinAndID_Response>> => {
  log("INFO", "API - Settings", "Media Server", "Fetching Plex PIN code and ID from backend");
  try {
    const response = await apiClient.get<APIResponse<GetPlexPinAndID_Response>>(`/oauth/plex`);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error fetching Plex PIN code and ID from backend");
    } else {
      log(
        "INFO",
        "API - Settings",
        "Media Server",
        "Fetched Plex PIN code and ID from backend successfully",
        response.data
      );
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Settings",
      "Media Server",
      `Failed to fetch Plex PIN code and ID from backend: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    const errorResponse = ReturnErrorMessage<GetPlexPinAndID_Response>(error);
    toast.error(errorResponse.error?.message || "Couldn't fetch Plex Pin Code", {
      duration: 1000,
    });
    return errorResponse;
  }
};

export interface PlexServerConnection {
  protocol: string;
  address: string;
  port: number;
  uri: string;
  local: boolean;
  relay: boolean;
  ipv6: boolean;
}

export interface PlexServersResponse {
  name: string;
  owned: boolean;
  connections: PlexServerConnection[];
}

export interface CheckAuthStatusWithPlex_Response {
  authenticated: boolean;
  auth_token: string;
  connections_available: PlexServersResponse[];
}

export const CheckAuthStatusWithPlex = async (
  plex_id: string
): Promise<APIResponse<CheckAuthStatusWithPlex_Response>> => {
  log("INFO", "API - Settings", "Media Server", "Checking authentication status with Plex");
  try {
    const params = { plex_id: plex_id };
    const response = await apiClient.post<APIResponse<CheckAuthStatusWithPlex_Response>>(`/oauth/plex`, null, {
      params: params,
    });
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Failed to check authentication status with Plex");
    } else {
      log(
        "INFO",
        "API - Settings",
        "Media Server",
        "Checked authentication status with Plex successfully",
        response.data
      );
    }

    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Settings",
      "Media Server",
      `Failed to check authentication status with Plex: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<CheckAuthStatusWithPlex_Response>(error);
  }
};
