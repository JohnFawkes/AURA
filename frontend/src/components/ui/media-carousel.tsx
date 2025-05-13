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
import { ZoomInIcon } from "lucide-react";
import { useRouter } from "next/navigation";
import { usePosterMediaStore } from "@/lib/setStore";
import { Lead, P } from "./typography";
import PosterSetModal from "./poster-set-modal";

type MediaCarouselProps = {
	set: PosterSet;
	mediaItem: MediaItem;
};

const formatDate = (dateString: string) => {
	try {
		const date = new Date(dateString);
		return new Intl.DateTimeFormat("en-US", {
			year: "numeric",
			month: "long",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit",
		}).format(date);
	} catch {
		return "Invalid Date";
	}
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
			<div className="flex flex-col">
				<div className="flex flex-row items-center">
					<P className="text-primary-dynamic">
						Set Author: {set.User.Name}
					</P>
					<div className="ml-auto flex space-x-2">
						<button
							className="btn"
							onClick={() => {
								goToSetPage();
							}}
						>
							<ZoomInIcon className="mb-1 mr-2 h-5 w-5 sm:h-7 sm:w-7" />
						</button>
						<button className="btn">
							<PosterSetModal
								posterSet={set}
								mediaItem={mediaItem}
							/>
						</button>
					</div>
				</div>
				<P className="text-sm text-muted-foreground flex items-center mb-1">
					Last Update: {formatDate(set.DateUpdated || "")}
				</P>
				<Lead
					className="text-sm text-muted-foreground"
					onClick={() => {
						goToSetPage();
					}}
				>
					{fileCounts}
				</Lead>
			</div>

			<CarouselContent>
				<ShowCarousel set={set as PosterSet} />
			</CarouselContent>
			<CarouselNext className="right-2 bottom-0" />
			<CarouselPrevious className="right-8 bottom-0" />
		</Carousel>
	);
}
