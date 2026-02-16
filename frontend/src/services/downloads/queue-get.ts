import apiClient from "@/services/api-client";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBSavedItem } from "@/types/database/db-poster-set";

export const getAllDownloadQueueItems = async (): Promise<
  APIResponse<{
    in_progress_entries: DBSavedItem[];
    error_entries: DBSavedItem[];
    warning_entries: DBSavedItem[];
  }>
> => {
  try {
    log("INFO", "API - Download Queue", "Fetch", "Fetching download queue entries");
    const response = await apiClient.get<
      APIResponse<{
        in_progress_entries: DBSavedItem[];
        error_entries: DBSavedItem[];
        warning_entries: DBSavedItem[];
      }>
    >(`/download/queue/item`);
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
    throw error;
  }
};
