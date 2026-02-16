import { FormItemDisplay } from "@/components/shared/download-modal";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { ImageFile, IncludedItem, SetRef } from "@/types/media-and-posters/sets";

const createBaseItem = (set: SetRef): FormItemDisplay => ({
	MediaItem: {} as MediaItem,
	Set: {
		id: set.id,
		title: set.title,
		type: set.type,
		user_created: set.user_created,
		date_created: set.date_created,
		date_updated: set.date_updated,
		popularity: set.popularity,
		popularity_global: set.popularity_global,
		images: [] as ImageFile[],
	},
});

export const setRefsToFormItems = (
	sets: SetRef[],
	includedItems: { [tmdb_id: string]: IncludedItem }
): FormItemDisplay[] => {
	const items: FormItemDisplay[] = [];
	sets.forEach((set) => {
		if (!set.item_ids || set.item_ids.length === 0 || !includedItems || !set.images || set.images.length === 0) {
			return; // Skip sets with no item_ids or images
		}

		// Get the MediaItem info from includedItems
		// If not found, don't append this item to the list
		// We can use the first item_id to look up the MediaItem
		// Then we can find the match images that have the same item_tmdb_id

		for (const itemID of set.item_ids) {
			const includedItem = includedItems[itemID];
			if (
				includedItem &&
				includedItem.media_item &&
				includedItem.media_item.rating_key &&
				includedItem.media_item.library_title
			) {
				const item = createBaseItem(set);
				item.MediaItem = includedItem.media_item;

				const matchedImages = set.images.filter(
					(image) => image.item_tmdb_id === includedItem.mediux_info.tmdb_id
				);

				if (matchedImages.length === 0) {
					continue; // Skip this item if no images match
				}

				item.Set.images = matchedImages;

				items.push(item);
			}
		}
	});

	return items;
};
