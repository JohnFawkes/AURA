import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { LibrarySection } from "@/types/media-and-posters/media-item-and-library";

export const getLibrarySectionItems = async (
  librarySection: LibrarySection,
  sectionStartIndex: number
): Promise<APIResponse<LibrarySection>> => {
  const logMessage =
    sectionStartIndex === 0
      ? `Fetching items for '${librarySection.title}'...`
      : `Fetching items for '${librarySection.title}' (index: ${sectionStartIndex})`;
  log("INFO", "API - Media Server", "Fetch Section Items", logMessage);
  try {
    const response = await apiClient.get<APIResponse<LibrarySection>>(`/mediaserver/library/items`, {
      params: {
        section_id: librarySection.id,
        section_title: librarySection.title,
        section_type: librarySection.type,
        section_start_index: sectionStartIndex,
      },
    });
    if (response.data.status === "error") {
      throw new Error(
        response.data.error?.message || `Unknown error fetching items for section '${librarySection.title}'`
      );
    } else {
      log("INFO", "API - Media Server", "Fetch Section Items", `Fetched items for '${librarySection.title}'`);
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Media Server",
      "Fetch Section Items",
      `Failed to fetch items for '${librarySection.title}': ${error instanceof Error ? error.message : "Unknown error"}`
    );
    return ReturnErrorMessage<LibrarySection>(error);
  }
};
