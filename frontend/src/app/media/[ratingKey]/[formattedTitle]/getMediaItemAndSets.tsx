import { MediaItem } from "@/types/mediaItem";
import { PosterSets } from "@/types/posterSets";
import { getMediaItemData } from "./getMediaItemData";
import { getPosterSets } from "./getPosterSets";

export async function getMediaItemAndSets(ratingKey: string): Promise<{
	mediaItem: MediaItem;
	posterSets: PosterSets;
}> {
	const mediaItemData = await getMediaItemData(ratingKey);
	const posterSet = await getPosterSets(mediaItemData.mediaItem);

	return {
		mediaItem: mediaItemData.mediaItem,
		posterSets: posterSet.posterSets,
	};
}
