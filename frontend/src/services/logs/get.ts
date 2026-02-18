import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse, LogData } from "@/types/api/api-response";

export interface GetLogContents_Response {
  possible_actions_paths: Record<string, { label: string; section: string }>;
  log_entries: LogData[];
  total_log_entries: number;
}

export const GetLogContents = async (
  filteredLogLevels: string[],
  filteredStatuses: string[],
  filteredActions: string[],
  itemsPerPage: number,
  pageNumber: number
): Promise<APIResponse<GetLogContents_Response>> => {
  log("INFO", "API - Logs", "Fetch Log Contents", "Fetching log contents");
  try {
    const params = {
      log_levels: filteredLogLevels.join(","),
      statuses: filteredStatuses.join(","),
      actions: filteredActions.join(","),
      items_per_page: itemsPerPage,
      page_number: pageNumber,
    };
    const response = await apiClient.get<APIResponse<GetLogContents_Response>>(`/logs`, { params: params });
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error fetching log contents");
    } else {
      log("INFO", "API - Logs", "Fetch Log Contents", "Fetched log contents successfully", response.data);
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Logs",
      "Fetch Log Contents",
      `Failed to fetch log contents: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<GetLogContents_Response>(error);
  }
};
