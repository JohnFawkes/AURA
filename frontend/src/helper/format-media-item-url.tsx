import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export function formatMediaItemUrl(mediaItem: MediaItem): string {
	const formattedTitle = mediaItem.Title.replace(/\s+/g, "_");
	const sanitizedTitle = formattedTitle.replace(/[^a-zA-Z0-9_]/g, "");
	return `/media/${mediaItem.RatingKey}/${sanitizedTitle}`;
}
