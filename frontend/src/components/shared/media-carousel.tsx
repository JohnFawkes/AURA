import { formatLastUpdatedDate } from "@/helper/formatDate";
import { ZoomInIcon } from "lucide-react";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { CarouselDisplay } from "@/components/shared/carousel-display";
import DownloadModal from "@/components/shared/download-modal";
import { SetFileCounts } from "@/components/shared/set-file-counts";
import { Carousel, CarouselContent, CarouselNext, CarouselPrevious } from "@/components/ui/carousel";
import { Lead } from "@/components/ui/typography";

import { useMediaStore } from "@/lib/mediaStore";
import { usePosterSetsStore } from "@/lib/posterSetStore";

import { MediaItem } from "@/types/mediaItem";
import { PosterSet } from "@/types/posterSets";

type MediaCarouselProps = {
	set: PosterSet;
	mediaItem: MediaItem;
};

export function MediaCarousel({ set, mediaItem }: MediaCarouselProps) {
	const router = useRouter();

	const { setPosterSets, setSetAuthor, setSetID, setSetTitle, setSetType } = usePosterSetsStore();
	const { setMediaItem } = useMediaStore();
	const goToSetPage = () => {
		setPosterSets([set]);
		setSetType(set.Type);
		setSetTitle(set.Title);
		setSetAuthor(set.User.Name);
		setSetID(set.ID);
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
					<div className="flex flex-row items-center">
						<Link
							href={`/sets/${set.ID}`}
							className="text-primary-dynamic hover:text-primary cursor-pointer text-md font-semibold"
							onClick={(e) => {
								e.stopPropagation();
								goToSetPage();
							}}
						>
							{set.Title}
						</Link>
					</div>
					<div className="ml-auto flex space-x-2">
						<Link
							href={`/sets/${set.ID}`}
							className="btn"
							onClick={(e) => {
								e.stopPropagation();
								goToSetPage();
							}}
						>
							<ZoomInIcon className="mr-2 h-5 w-5 sm:h-7 sm:w-7 cursor-pointer" />
						</Link>
						<DownloadModal
							setType={set.Type}
							setTitle={set.Title}
							setID={set.ID}
							setAuthor={set.User.Name}
							posterSets={[set]}
						/>
					</div>
				</div>
				<div className="text-md text-muted-foreground  mb-1">
					By:{" "}
					<Link href={`/user/${set.User.Name}`} className="hover:text-primary cursor-pointer">
						{set.User.Name}
					</Link>
				</div>
				<Lead className="text-sm text-muted-foreground flex items-center mb-1">
					Last Update: {formatLastUpdatedDate(set.DateUpdated, set.DateCreated)}
				</Lead>

				<SetFileCounts mediaItem={mediaItem} set={set} />
			</div>

			<CarouselContent>
				<CarouselDisplay sets={[set] as PosterSet[]} />
			</CarouselContent>
			<CarouselNext className="right-2 bottom-0" />
			<CarouselPrevious className="right-8 bottom-0" />
		</Carousel>
	);
}
