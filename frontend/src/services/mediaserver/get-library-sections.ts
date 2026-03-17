import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { LibrarySection } from "@/types/media-and-posters/media-item-and-library";

export interface GetLibrarySections_Response {
  sections: LibrarySection[];
}

export const GetLibrarySections = async (): Promise<APIResponse<GetLibrarySections_Response>> => {
  log("INFO", "API - Media Server", "Fetch Library Sections", "Fetching all library sections");
  try {
    const response = await apiClient.get<APIResponse<GetLibrarySections_Response>>(`/mediaserver/libraries/`);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error fetching all library sections");
    } else {
      log("INFO", "API - Media Server", "Fetch Library Sections", `Fetched all library sections successfully`);
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Media Server",
      "Fetch Library Sections",
      `Failed to fetch all library sections: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<GetLibrarySections_Response>(error);
  }
};
