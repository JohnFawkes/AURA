import { CollectionItem } from "@/app/collections/page";
import { collectionItemInfo } from "@/helper/item-info";
import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import { APIResponse } from "@/types/api/api-response";
import { ImageFile } from "@/types/media-and-posters/sets";

export const downloadImageFileForCollectionItem = async (
  imageType: "poster" | "backdrop",
  collectionItem: CollectionItem,
  imageFile: ImageFile
): Promise<APIResponse<string>> => {
  log(
    "INFO",
    "API - Media Server",
    "Download and Update Collection Image",
    `Downloading ${imageType} image and updating ${collectionItemInfo(collectionItem)}`
  );
  try {
    const response = await apiClient.post<APIResponse<string>>(`/download/image/collection`, {
      collection_item: collectionItem,
      image_file: imageFile,
    });
    if (response.data.status === "error") {
      throw new Error(
        response.data.error?.message || `Unknown error downloading ${imageType} image and updating media server`
      );
    } else {
      log(
        "INFO",
        "API - Media Server",
        "Download and Update Collection Image",
        `Downloaded ${imageType} image and updated ${collectionItemInfo(collectionItem)}`,
        response.data
      );
    }
    return response.data;
  } catch (error) {
    log(
      "ERROR",
      "API - Media Server",
      "Download and Update Collection Image",
      `Failed to download ${imageType} image and update ${collectionItemInfo(collectionItem)}
			${error instanceof Error ? error.message : "Unknown error"}`,
      error
    );
    return ReturnErrorMessage<string>(error);
  }
};
