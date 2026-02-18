import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBSavedItem } from "@/types/database/db-poster-set";

export interface GetAllDownloadQueueItems_Response {
  in_progress_entries: DBSavedItem[];
  warning_entries: DBSavedItem[];
  error_entries: DBSavedItem[];
}

export const GetAllDownloadQueueItems = async (): Promise<APIResponse<GetAllDownloadQueueItems_Response>> => {
  try {
    log("INFO", "API - Download Queue", "Fetch", "Fetching download queue entries");
    const response = await apiClient.get<APIResponse<GetAllDownloadQueueItems_Response>>(`/download/queue/item`);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error fetching download queue entries");
    } else {
      log("INFO", "API - Download Queue", "Fetch", "Fetched download queue entries successfully", {
        in_progress_entries: response.data.data?.in_progress_entries,
        error_entries: response.data.data?.error_entries,
        warning_entries: response.data.data?.warning_entries,
      });
    }
    return response.data;
  } catch (error) {
    log("ERROR", "API - Download Queue", "Fetch", "Error fetching download queue entries", { error });
    return ReturnErrorMessage<GetAllDownloadQueueItems_Response>(error);
  }
};
