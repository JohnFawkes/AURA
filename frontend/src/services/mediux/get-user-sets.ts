import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { CreatorSetsResponse } from "@/types/media-and-posters/sets";

export interface GetAllUserSets_Response {
  sets: CreatorSetsResponse;
}

export const GetAllUserSets = async (username: string): Promise<APIResponse<GetAllUserSets_Response>> => {
  log("INFO", "API - MediUX", "Fetch All User Sets", `Fetching all user sets for ${username}`);
  try {
    const params = {
      username: username,
    };
    const response = await apiClient.get<APIResponse<GetAllUserSets_Response>>("/mediux/sets/user", { params });
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || `Unknown error fetching all user sets for ${username}`);
    } else {
      log(
        "INFO",
        "API - MediUX",
        "Fetch All User Sets",
        `Fetched all user sets for ${username} successfully`,
        response.data
      );
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - MediUX",
      "Fetch All User Sets",
      `Failed to fetch all user sets for ${username}: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<GetAllUserSets_Response>(error);
  }
};
