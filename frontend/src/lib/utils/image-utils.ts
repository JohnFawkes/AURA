/**
 * Image utility functions for handling aspect ratios and sizing
 */
export type AspectRatio = "poster" | "backdrop" | "titlecard" | "album" | "logo";

/**
 * Returns the appropriate Tailwind class for a given aspect ratio
 */
export function getAspectRatioClass(aspect: AspectRatio): string {
	switch (aspect) {
		case "poster":
			return "aspect-[2/3]";
		case "backdrop":
		case "titlecard":
		case "logo":
			return "aspect-video";
		case "album":
			return "aspect-square";
		default:
			return "aspect-[2/3]";
	}
}

/**
 * Returns the appropriate sizes attribute for image elements
 */
export function getImageSizes(type: string): string {
	return ["backdrop", "titlecard"].includes(type.toLowerCase()) ? "(max-width: 640px) 50vw, 300px" : "300px";
}
