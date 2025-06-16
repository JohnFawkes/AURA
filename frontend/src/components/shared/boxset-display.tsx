import {
	BoxsetCollectionToPosterSet,
	BoxsetMovieToPosterSet,
	BoxsetShowToPosterSet,
	BoxsetToPosterSet,
} from "@/helper/boxsetToPosterSet";
import { formatLastUpdatedDate } from "@/helper/formatDate";
import { CheckCircle2 as Checkmark } from "lucide-react";

import { CarouselDisplay } from "@/components/shared/carousel-display";
import DownloadModal from "@/components/shared/download-modal";
import { ShowFullSetsDisplay } from "@/components/shared/show-full-set";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

import {
	MediuxUserBoxset,
	MediuxUserCollectionSet,
	MediuxUserMovieSet,
	MediuxUserShowSet,
} from "@/types/mediuxUserAllSets";
import { PosterSet } from "@/types/posterSets";

import { Carousel, CarouselContent, CarouselNext, CarouselPrevious } from "../ui/carousel";
import { Lead } from "../ui/typography";

export const RenderBoxSetDisplay = ({
	set,
	type,
}: {
	set: MediuxUserShowSet | MediuxUserMovieSet | MediuxUserCollectionSet | MediuxUserBoxset;
	type: "show" | "movie" | "collection" | "boxset";
}) => {
	// Handle different types of sets
	const getPosterSets = () => {
		switch (type) {
			case "show":
				return BoxsetShowToPosterSet(set as MediuxUserShowSet);
			case "movie":
				return BoxsetMovieToPosterSet(set as MediuxUserMovieSet);
			case "collection":
				return BoxsetCollectionToPosterSet(set as MediuxUserCollectionSet);
			case "boxset":
				return BoxsetToPosterSet(set as MediuxUserBoxset);
			default:
				return [];
		}
	};

	const posterSets = getPosterSets();
	if (!posterSets || posterSets.length === 0) return null;

	// For boxsets, render with accordion
	if (type === "boxset") {
		const boxset = set as MediuxUserBoxset;
		return (
			<Accordion type="single" collapsible className="w-full">
				<AccordionItem value={boxset.id}>
					<AccordionTrigger className="flex items-center justify-between">
						<div className="text-primary-dynamic hover:text-primary cursor-pointer text-lg font-semibold">
							{boxset.boxset_title}
						</div>
					</AccordionTrigger>
					<AccordionContent>
						<div className="flex justify-end">
							<DownloadModal
								setType={type}
								setTitle={boxset.boxset_title}
								setID={boxset.id}
								setAuthor={boxset.user_created.username}
								posterSets={posterSets}
							/>
						</div>
						<ShowFullSetsDisplay
							setType={type}
							setTitle={boxset.boxset_title}
							setAuthor={boxset.user_created.username}
							setID={boxset.id}
							posterSets={posterSets}
						/>
					</AccordionContent>
				</AccordionItem>
			</Accordion>
		);
	}

	// For other types, render carousel
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
						<div className="text-primary-dynamic hover:text-primary cursor-pointer text-md font-semibold">
							{type === "collection"
								? (set as MediuxUserCollectionSet).set_title
								: (set as MediuxUserShowSet | MediuxUserMovieSet).set_title}
						</div>
						{type === "movie" && (set as MediuxUserMovieSet).movie_id.MediaItem?.ExistInDatabase && (
							<Checkmark className="ml-2 text-green-500" size={20} />
						)}
						{type === "show" && (set as MediuxUserShowSet).show_id.MediaItem?.ExistInDatabase && (
							<Checkmark className="ml-2 text-green-500" size={20} />
						)}
					</div>
					<div className="ml-auto flex space-x-2">
						<DownloadModal
							setType={type}
							setTitle={
								type === "collection"
									? (set as MediuxUserCollectionSet).set_title
									: (set as MediuxUserShowSet | MediuxUserMovieSet).set_title
							}
							setID={set.id}
							setAuthor={set.user_created.username}
							posterSets={posterSets}
						/>
					</div>
				</div>
				<Lead className="text-sm text-muted-foreground flex items-center mb-1">
					Last Update: {formatLastUpdatedDate(set.date_updated, set.date_created)}
				</Lead>
			</div>

			<CarouselContent>
				<CarouselDisplay sets={posterSets as PosterSet[]} />
			</CarouselContent>
			<CarouselNext className="right-2 bottom-0" />
			<CarouselPrevious className="right-8 bottom-0" />
		</Carousel>
	);
};
