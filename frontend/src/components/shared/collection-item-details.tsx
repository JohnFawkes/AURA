"use client";

import { CollectionItem } from "@/app/collections/page";

import { AssetImage } from "@/components/shared/asset-image";
import { CollectionMediaItemsCarousel } from "@/components/shared/collection-media-items-carousel";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";
import { H1, Lead } from "@/components/ui/typography";

type CollectionItemDetailsProps = {
	collectionItem: CollectionItem;
};

export function CollectionItemDetails({ collectionItem }: CollectionItemDetailsProps) {
	const minYear = collectionItem.MediaItems.reduce((min, item) => (item.Year < min ? item.Year : min), Infinity);
	const maxYear = collectionItem.MediaItems.reduce((max, item) => (item.Year > max ? item.Year : max), -Infinity);

	return (
		<div>
			<div className="flex flex-col lg:flex-row pt-30 items-center lg:items-start text-center lg:text-left">
				{/* Poster Image */}

				<div className="flex-shrink-0 mb-4 lg:mb-0 lg:mr-8 flex justify-center">
					<AssetImage
						image={`/api/mediaserver/image?ratingKey=${collectionItem?.RatingKey}&imageType=poster&cb=${Date.now()}`}
						className="w-[200px] h-auto transition-transform hover:scale-105 select-none"
					/>
				</div>

				{/* Title and Summary */}
				<div className="flex flex-col items-center lg:items-start">
					<H1 className="mb-1">{collectionItem?.Title}</H1>
					{/* Hide summary on mobile */}
					<Lead className="text-primary-dynamic max-w-xl hidden lg:block">{collectionItem.Summary}</Lead>
					<Badge variant="default" className="mt-2 mb-2">
						{minYear === maxYear ? `${minYear}` : `${minYear} - ${maxYear}`}
					</Badge>
				</div>
			</div>

			{/* Library Information */}
			{collectionItem.LibraryTitle && (
				<div className="flex flex-wrap lg:flex-nowrap justify-center lg:justify-start items-center gap-4 tracking-wide mt-0 md:mt-2">
					<Lead className="text-md text-primary-dynamic ml-1">
						<span className="font-semibold">{collectionItem.LibraryTitle} Library</span>
					</Lead>
				</div>
			)}

			{/* Loop through the Media Items and display their posters */}
			{collectionItem.MediaItems && collectionItem.MediaItems.length > 0 && (
				<Accordion type="single" collapsible className="w-full">
					<AccordionItem value="media-items">
						<AccordionTrigger className="text-primary font-semibold">
							View {collectionItem.MediaItems.length} Media Items
						</AccordionTrigger>
						<AccordionContent>
							<CollectionMediaItemsCarousel mediaItems={collectionItem.MediaItems} />
						</AccordionContent>
					</AccordionItem>
				</Accordion>
			)}
		</div>
	);
}
