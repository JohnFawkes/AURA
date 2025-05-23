"use client";

import {
	Carousel,
	CarouselContent,
	CarouselNext,
	CarouselPrevious,
} from "@/components/ui/carousel";

import { PosterSet } from "@/types/posterSets";
import { CarouselShow } from "../carousel-show";
import { CarouselMovie } from "../carousel-movie";
import { MediaItem } from "@/types/mediaItem";
import { ZoomInIcon } from "lucide-react";
import { useRouter } from "next/navigation";
import { Lead, P } from "./typography";
import { usePosterSetStore } from "@/lib/posterSetStore";
import { useMediaStore } from "@/lib/mediaStore";
import { formatLastUpdatedDate } from "@/helper/formatDate";
import { SetFileCounts } from "../set_file_counts";
import DownloadModalShow from "../download-modal-show";
import DownloadModalMovie from "../download-modal-movie";

type MediaCarouselProps = {
	set: PosterSet;
	mediaItem: MediaItem;
};

export function MediaCarousel({ set, mediaItem }: MediaCarouselProps) {
	const router = useRouter();

	const { setPosterSet } = usePosterSetStore();
	const { setMediaItem } = useMediaStore();

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
					<P
						className="text-primary-dynamic hover:text-primary cursor-pointer text-md font-semibold"
						onClick={() => {
							goToSetPage();
						}}
					>
						{set.Title} by {set.User.Name}
					</P>
					<div className="ml-auto flex space-x-2">
						<button
							className="btn"
							onClick={() => {
								goToSetPage();
							}}
						>
							<ZoomInIcon className="mr-2 h-5 w-5 sm:h-7 sm:w-7" />
						</button>
						{mediaItem.Type === "show" ? (
							<button className="btn">
								<DownloadModalShow
									posterSet={set}
									mediaItem={mediaItem}
								/>
							</button>
						) : mediaItem.Type === "movie" ? (
							<button className="btn">
								<DownloadModalMovie
									posterSet={set}
									mediaItem={mediaItem}
								/>
							</button>
						) : null}
					</div>
				</div>
				<Lead className="text-sm text-muted-foreground flex items-center mb-1">
					Last Update:{" "}
					{formatLastUpdatedDate(set.DateUpdated, set.DateCreated)}
				</Lead>

				<SetFileCounts mediaItem={mediaItem} set={set} />
			</div>

			<CarouselContent>
				{mediaItem.Type === "show" ? (
					<CarouselShow set={set as PosterSet} />
				) : mediaItem.Type === "movie" ? (
					<CarouselMovie set={set as PosterSet} />
				) : null}
			</CarouselContent>
			<CarouselNext className="right-2 bottom-0" />
			<CarouselPrevious className="right-8 bottom-0" />
		</Carousel>
	);
}
