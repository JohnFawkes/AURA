import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";
import { toast } from "sonner";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";

export const getPlexPinCodeAndIDFromBackend = async (): Promise<{ ok: boolean; plex_pin: string; plex_id: string }> => {
  log("INFO", "API - Settings", "Media Server", "Fetching Plex PIN code and ID from backend");
  try {
    const response = await apiClient.get<APIResponse<{ plex_pin: string; plex_id: string }>>(`/oauth/plex`);
    if (response.data.status === "error") {
      return {
        ok: false,
        plex_pin: "",
        plex_id: "",
      };
    }
    log(
      "INFO",
      "API - Settings",
      "Media Server",
      "Fetched Plex PIN code and ID from backend successfully",
      response.data
    );
    return { ok: true, plex_pin: response.data.data?.plex_pin || "", plex_id: response.data.data?.plex_id || "" };
  } catch (error) {
    log(
      "ERROR",
      "API - Settings",
      "Media Server",
      `Failed to fetch Plex PIN code and ID from backend: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    const errorResponse = ReturnErrorMessage<{ plex_pin: string; plex_id: string }>(error);
    toast.error(errorResponse.error?.message || "Couldn't fetch Plex Pin Code", {
      duration: 1000,
    });
    return {
      ok: false,
      plex_pin: "",
      plex_id: "",
    };
  }
};

export const checkAuthStatusWithPlex = async (
  plex_id: string
): Promise<{
  ok: boolean;
  authenticated: boolean;
  auth_token: string;
  connections_available: PlexServersResponse[];
}> => {
  log("INFO", "API - Settings", "Media Server", "Checking authentication status with Plex");
  try {
    const response = await apiClient.post<
      APIResponse<{ authenticated: boolean; auth_token: string; connections_available: PlexServersResponse[] }>
    >(`/oauth/plex`, null, {
      params: { plex_id: plex_id },
    });
    if (response.data.status === "error") {
      return {
        ok: false,
        authenticated: false,
        auth_token: "",
        connections_available: [],
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
      auth_token: response.data.data?.auth_token || "",
      connections_available: response.data.data?.connections_available || [],
    };
  } catch {
    log("ERROR", "API - Settings", "Media Server", `Failed to check authentication status with Plex`);
    return {
      ok: false,
      authenticated: false,
      auth_token: "",
      connections_available: [],
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
