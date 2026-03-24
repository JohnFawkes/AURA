import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { DBSavedItem } from "@/types/database/db-poster-set";

export interface RemoveItemFromQueue_Request {
  item: DBSavedItem;
}
export interface RemoveItemFromQueue_Response {
  result: string;
}

export const RemoveItemFromQueue = async (dbItem: DBSavedItem): Promise<APIResponse<RemoveItemFromQueue_Response>> => {
  log(
    "INFO",
    "API - Media Server",
    "Delete from Queue",
    `Deleting '${dbItem.media_item?.title ?? "(unknown)"}' (TMDB ID: ${dbItem.media_item?.tmdb_id ?? "(unknown)"}) from the download queue`
  );
  if (Array.isArray(dbItem.poster_sets) && dbItem.poster_sets.length > 0) {
    for (const set of dbItem.poster_sets) {
      set.last_downloaded = new Date().toISOString();
    }
  }
  try {
    const req: RemoveItemFromQueue_Request = {
      item: dbItem,
    };
    const response = await apiClient.delete<APIResponse<RemoveItemFromQueue_Response>>(`/download/queue/item`, {
      data: req,
    });
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error while deleting from download queue");
    } else {
      log(
        "INFO",
        "API - Media Server",
        "Delete from Queue",
        `Deleted '${dbItem.media_item.title}' (TMDB ID: ${dbItem.media_item.tmdb_id}) from the download queue`,
        response.data
      );
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Media Server",
      "Delete from Queue",
      `Failed to delete '${dbItem.media_item.title}' (TMDB ID: ${dbItem.media_item.tmdb_id}) from the download queue: ${
        dbItem.media_item.rating_key
      }: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<RemoveItemFromQueue_Response>(error);
  }
};
