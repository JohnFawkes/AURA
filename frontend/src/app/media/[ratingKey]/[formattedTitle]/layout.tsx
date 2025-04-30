import { extractAndProcessColors } from "@/lib/color-extraction";
import { getMediaItemAndSets } from "./getMediaItemAndSets";
import { ContentBackground } from "@/components/ui/backgrounds/content-background";

type Props = {
	children: React.ReactNode;
	params: {
		ratingKey: string;
	};
};

export default async function MediaLayout({ children, params }: Props) {
	const resolvedParams = await params;
	//const { ratingKey } = resolvedParams;

	// Fetch the media item data and poster sets data
	//const { mediaItem } = await getMediaItemAndSets(ratingKey);

	// let dynamicPalette = undefined;
	// if (ratingKey) {
	// 	try {
	// 		const backdropUrl = `/api/mediaserver/image/${ratingKey}/backdrop`;
	// 		console.log(backdropUrl);

	// 		// Extra check for invalid URL patterns
	// 		if (
	// 			!backdropUrl.includes("NULL") &&
	// 			!backdropUrl.includes("undefined")
	// 		) {
	// 			console.log(
	// 				`Requesting colors for ${ratingKey} with backdrop: ${backdropUrl}`
	// 			);

	// 			const startTime = Date.now();
	// 			dynamicPalette = await extractAndProcessColors(backdropUrl);
	// 			const endTime = Date.now();

	// 			console.log(
	// 				`Color extraction for ${ratingKey} completed in ${
	// 					endTime - startTime
	// 				}ms`
	// 			);
	// 		} else {
	// 			console.log(
	// 				`Invalid backdrop URL for ${ratingKey}: ${backdropUrl}`
	// 			);
	// 		}
	// 	} catch (error) {
	// 		console.warn("Failed to extract colors:", error);
	// 	}
	// } else {
	// 	console.log(`No valid backdrop found for ${ratingKey}`);
	// }

	return <div className="relative">{children}</div>;
}
