import { formatLastUpdatedDate } from "@/helper/format-date-last-updates";
import { Database, User, ZoomInIcon } from "lucide-react";

import Link from "next/link";
import { useRouter } from "next/navigation";

import { CarouselDisplay } from "@/components/shared/carousel-display";
import DownloadModal from "@/components/shared/download-modal";
import { SetFileCounts } from "@/components/shared/set-file-counts";
import { Carousel, CarouselContent, CarouselNext, CarouselPrevious } from "@/components/ui/carousel";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Lead } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { useMediaStore } from "@/lib/stores/global-store-media-store";
import { usePosterSetsStore } from "@/lib/stores/global-store-poster-sets";
import { useSearchQueryStore } from "@/lib/stores/global-store-search-query";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";
import { PosterSet } from "@/types/media-and-posters/poster-sets";

type MediaCarouselProps = {
	set: PosterSet;
	mediaItem: MediaItem;
	onMediaItemChange?: (item: MediaItem) => void;
};

export function MediaCarousel({ set, mediaItem, onMediaItemChange }: MediaCarouselProps) {
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

	const { setSearchQuery } = useSearchQueryStore();

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
					<div className="flex flex-row items-center gap-2">
						<Link
							href={`/sets/${set.ID}`}
							className="text-primary-dynamic hover:text-primary cursor-pointer text-md font-semibold ml-1"
							onClick={(e) => {
								e.stopPropagation();
								goToSetPage();
							}}
						>
							{set.Title}
						</Link>
						{mediaItem &&
							mediaItem.DBSavedSets &&
							mediaItem.DBSavedSets.map((s) => s.PosterSetID).includes(set.ID) && (
								<Popover>
									<PopoverTrigger asChild>
										<Database
											className="text-green-500 hover:text-green-600 cursor-pointer active:scale-95"
											size={20}
										/>
									</PopoverTrigger>
									<PopoverContent
										side="top"
										sideOffset={5}
										className="bg-secondary border border-2 border-primary p-2"
									>
										<div className={cn("text-center", "text-md", "font-medium", "rounded-md")}>
											<Database
												className="inline-block mr-2 text-green-500 hover:text-green-600 cursor-pointer active:scale-95"
												size={16}
												onClick={(e) => {
													e.stopPropagation();
													setSearchQuery(
														`${mediaItem.Title} Y:${mediaItem.Year}: ID:${mediaItem.TMDB_ID}: L:${mediaItem.LibraryTitle}:`
													);
													router.push("/saved-sets");
												}}
											/>
											This exact set is already in your database
										</div>
									</PopoverContent>
								</Popover>
							)}
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
							<ZoomInIcon className="mr-2 h-5 w-5 sm:h-7 sm:w-7 cursor-pointer active:scale-95 hover:text-primary" />
						</Link>
						<DownloadModal
							setType={set.Type}
							setTitle={set.Title}
							setID={set.ID}
							setAuthor={set.User.Name}
							posterSets={[set]}
							onMediaItemChange={onMediaItemChange}
						/>
					</div>
				</div>
				<div className="text-md text-muted-foreground mb-1 flex items-center">
					<User />
					<Link href={`/user/${set.User.Name}`} className="hover:text-primary cursor-pointer underline">
						{set.User.Name}
					</Link>
				</div>
				<Lead className="text-sm text-muted-foreground flex items-center mb-1 ml-1">
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
