import { fetchMediaServerItemContent } from "@/services/api.mediaserver";
import { MediaItem } from "@/types/mediaItem";

export async function getMediaItemData(ratingKey: string): Promise<{
	mediaItem: MediaItem;
}> {
	try {
		console.log("INSIDE LAYOUT");
		const resp = await fetchMediaServerItemContent(ratingKey);
		if (!resp) {
			throw new Error("No response from API");
		}
		if (resp.status !== "success") {
			throw new Error(resp.message);
		}
		const responseItem = resp.data;
		if (!responseItem) {
			throw new Error("No item found in response");
		}
		return { mediaItem: responseItem };
	} catch (error) {
		if (error instanceof Error) {
			console.error("Error fetching Media Item:", error.message);
			return Promise.reject(error.message);
		} else {
			console.error("An unknown error occurred:", error);
			return Promise.reject("An unknown error occurred, check logs.");
		}
	}
}
