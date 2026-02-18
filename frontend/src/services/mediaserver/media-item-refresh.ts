import { mediaItemInfo } from "@/helper/item-info";
import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export interface RefreshMediaItemMetadata_Response {
  refreshed: boolean;
}

export const RefreshMediaItemMetadata = async (
  mediaItem: MediaItem,
  refreshTitle: string,
  refreshKey: string
): Promise<APIResponse<RefreshMediaItemMetadata_Response>> => {
  log(
    "INFO",
    "API - Media Server",
    "Refresh Item",
    `Refreshing ${mediaItemInfo(mediaItem)} - ${refreshTitle} (Key: ${refreshKey})`
  );
  try {
    const params = {
      rating_key: mediaItem.rating_key,
      refresh_rating_key: refreshKey,
    };
    const response = await apiClient.post<APIResponse<RefreshMediaItemMetadata_Response>>(
      `/mediaserver/refresh`,
      null,
      { params }
    );
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error refreshing media server item");
    } else {
      log(
        "INFO",
        "API - Media Server",
        "Refresh Item",
        `Refreshed media item ${mediaItemInfo(mediaItem)} - ${refreshTitle}`,
        response.data
      );
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Media Server",
      "Refresh Item",
      `Failed to refresh media item ${mediaItemInfo(mediaItem)} - ${refreshTitle} (Key: ${refreshKey}) ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<RefreshMediaItemMetadata_Response>(error);
  }
};
