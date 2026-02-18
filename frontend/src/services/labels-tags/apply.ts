import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { DBSavedItem } from "@/types/database/db-poster-set";
import { MediaItem, SelectedTypes } from "@/types/media-and-posters/media-item-and-library";

export interface ApplyLabelsAndTagsToItem_Request {
  media_item: MediaItem;
  selected_types: SelectedTypes;
}

export interface ApplyLabelsAndTagsToItem_Response {
  message: string;
}

export const ApplyLabelsAndTagsToItem = async (
  dbItem: DBSavedItem
): Promise<APIResponse<ApplyLabelsAndTagsToItem_Response>> => {
  log(
    "INFO",
    "API - Labels/Tags",
    "Apply Labels/Tags",
    `Applying labels/tags to '${dbItem.media_item.title}' (TMDB ID: ${dbItem.media_item.tmdb_id})`
  );

  const selectedTypes: SelectedTypes = {
    poster: false,
    backdrop: false,
    season_poster: false,
    special_season_poster: false,
    titlecard: false,
  };
  for (const posterSet of dbItem.poster_sets) {
    selectedTypes.poster = selectedTypes.poster || posterSet.selected_types.poster;
    selectedTypes.backdrop = selectedTypes.backdrop || posterSet.selected_types.backdrop;
    selectedTypes.season_poster = selectedTypes.season_poster || posterSet.selected_types.season_poster;
    selectedTypes.special_season_poster =
      selectedTypes.special_season_poster || posterSet.selected_types.special_season_poster;
    selectedTypes.titlecard = selectedTypes.titlecard || posterSet.selected_types.titlecard;
  }

  try {
    const req: ApplyLabelsAndTagsToItem_Request = {
      media_item: dbItem.media_item,
      selected_types: selectedTypes,
    };
    const response = await apiClient.post<APIResponse<ApplyLabelsAndTagsToItem_Response>>(`/labels-tags`, req);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error while applying labels/tags");
    } else {
      log(
        "INFO",
        "API - Labels/Tags",
        "Apply Labels/Tags",
        `Applied labels/tags to '${dbItem.media_item.title}' (TMDB ID: ${dbItem.media_item.tmdb_id})`,
        response.data
      );
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Labels/Tags",
      "Apply Labels/Tags",
      `Failed to apply labels/tags to '${dbItem.media_item.title}' (TMDB ID: ${dbItem.media_item.tmdb_id}): ${
        dbItem.media_item.rating_key
      }: ${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<ApplyLabelsAndTagsToItem_Response>(error);
  }
};
