import { useRouter } from "next/navigation";

import { AssetImage } from "@/components/shared/asset-image";
import { Carousel, CarouselContent, CarouselItem, CarouselNext, CarouselPrevious } from "@/components/ui/carousel";

import { useMediaStore } from "@/lib/stores/global-store-media-store";

import { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export function CollectionMediaItemsCarousel({ mediaItems }: { mediaItems: MediaItem[] }) {
	const router = useRouter();

	const { setMediaItem } = useMediaStore();

	const handleCardClick = (mediaItem: MediaItem) => {
		setMediaItem(mediaItem);
		//router.push(formatMediaItemUrl(mediaItem));
		router.push("/media-item/");
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
			<CarouselContent>
				{mediaItems.map((mediaItem) => (
					<CarouselItem key={`${mediaItem.RatingKey}`} onClick={() => handleCardClick(mediaItem)}>
						<p className="text-primary-dynamic text-md text-center mt-2 mb-1">{mediaItem.Title}</p>
						<AssetImage image={mediaItem} aspect="poster" className={`w-full`} />
						<AssetImage image={mediaItem} aspect="backdrop" className={`w-full`} />
					</CarouselItem>
				))}
			</CarouselContent>
			<CarouselNext className="right-2 bottom-0" />
			<CarouselPrevious className="right-8 bottom-0" />
		</Carousel>
	);
}
