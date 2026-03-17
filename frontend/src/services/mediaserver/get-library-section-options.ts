import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { AppConfigMediaServer } from "@/types/config/config";
import type { LibrarySection } from "@/types/media-and-posters/media-item-and-library";

export interface GetLibrarySectionOptions_Request {
  media_server: AppConfigMediaServer;
}
export interface GetLibrarySectionOptions_Response {
  library_sections: LibrarySection[];
}

export const GetLibrarySectionOptions = async (
  config: AppConfigMediaServer
): Promise<APIResponse<GetLibrarySectionOptions_Response>> => {
  log("INFO", "API - Media Server", "Fetch Library Section Options", "Fetching all library section options");
  try {
    const req: GetLibrarySectionOptions_Request = {
      media_server: config,
    };
    const response = await apiClient.post<APIResponse<GetLibrarySectionOptions_Response>>(
      `/mediaserver/libraries/options`,
      req
    );
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error fetching all library section options");
    } else {
      log(
        "INFO",
        "API - Media Server",
        "Fetch Library Section Options",
        `Fetched all library section options successfully`
      );
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Media Server",
      "Fetch Library Section Options",
      `Failed to fetch all library section options: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<GetLibrarySectionOptions_Response>(error);
  }
};
