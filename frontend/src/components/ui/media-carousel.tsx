"use client";

import {
	Carousel,
	CarouselContent,
	CarouselNext,
	CarouselPrevious,
} from "@/components/ui/carousel";

import { PosterSet } from "@/types/posterSets";
import { ShowCarousel } from "./show-carousel";
import { MediaItem } from "@/types/mediaItem";
import { Download } from "lucide-react";
import { useRouter } from "next/navigation";
import { usePosterMediaStore } from "@/lib/setStore";

type MediaCarouselProps = {
	set: PosterSet;
	mediaItem: MediaItem;
};

export function MediaCarousel({ set, mediaItem }: MediaCarouselProps) {
	const router = useRouter();

	// Helper function to get the count and label for each type
	const getFileTypeCount = (type: string, label: string) => {
		const count = set.Files.filter((file) => file.Type === type).length;
		return count > 0 ? `${count} ${label}${count > 1 ? "s" : ""}` : null;
	};

	// Get counts for each type
	const posterCount = getFileTypeCount("poster", "Poster");
	const backdropCount = getFileTypeCount("backdrop", "Backdrop");
	const seasonPosterCount = getFileTypeCount("seasonPoster", "Season Poster");
	const titlecardCount = getFileTypeCount("titlecard", "Titlecard");

	// Combine counts into a single string
	const fileCounts = [
		posterCount,
		backdropCount,
		seasonPosterCount,
		titlecardCount,
	]
		.filter(Boolean) // Remove null values
		.join(" • ");

	const { setPosterSet, setMediaItem } = usePosterMediaStore();
	const goToSetPage = () => {
		setPosterSet(set);
		setMediaItem(mediaItem);
		router.push(`/sets/${set.ID}`);
	};

	return (
		<Carousel
			opts={{
				align: "start",
				dragFree: true,
				slidesToScroll: "auto",
			}}
			className="w-full"
		>
			<div className="flex flex-col ">
				<div className="text-xs flex flex-row items-center gap-1">
					<span className="text-muted-foreground">Set by:</span>
					<span className="text-primary-dynamic flex items-center gap-1">
						{set.User.Name}{" "}
					</span>
					<div className="ml-auto">
						<button
							className="btn"
							onClick={() => {
								goToSetPage();
							}}
						>
							<Download className="mr-2 h-4 w-4" />
						</button>
					</div>
				</div>
				<span
					className="text-xs text-muted-foreground"
					onClick={() => {
						goToSetPage();
					}}
				>
					{fileCounts}
				</span>
			</div>

			<CarouselContent>
				<ShowCarousel set={set as PosterSet} />
			</CarouselContent>
			<CarouselNext className="right-2 bottom-0" />
			<CarouselPrevious className="right-8 bottom-0" />
		</Carousel>
	);
}
