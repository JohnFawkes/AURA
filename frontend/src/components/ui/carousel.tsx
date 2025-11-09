"use client";

import useEmblaCarousel, { type UseEmblaCarouselType } from "embla-carousel-react";
import { ChevronLeft, ChevronRight } from "lucide-react";

import * as React from "react";

import { useViewDensity } from "@/components/shared/view-density-context";
import { Button } from "@/components/ui/button";

import { cn } from "@/lib/cn";

type CarouselApi = UseEmblaCarouselType[1];
type UseCarouselParameters = Parameters<typeof useEmblaCarousel>;
type CarouselOptions = UseCarouselParameters[0];
type CarouselPlugin = UseCarouselParameters[1];

type CarouselProps = {
	opts?: CarouselOptions;
	plugins?: CarouselPlugin;
	orientation?: "horizontal" | "vertical";
	setApi?: (api: CarouselApi) => void;
	useDensityScaling?: boolean;
};

type CarouselContextProps = {
	carouselRef: ReturnType<typeof useEmblaCarousel>[0];
	api: ReturnType<typeof useEmblaCarousel>[1];
	scrollPrev: () => void;
	scrollNext: () => void;
	canScrollPrev: boolean;
	canScrollNext: boolean;
	useDensityScaling: boolean;
} & CarouselProps;

const CarouselContext = React.createContext<CarouselContextProps | null>(null);

function useCarousel() {
	const context = React.useContext(CarouselContext);

	if (!context) {
		throw new Error("useCarousel must be used within a <Carousel />");
	}

	return context;
}

function Carousel({
	orientation = "horizontal",
	opts,
	setApi,
	plugins,
	className,
	useDensityScaling = true,
	children,
	...props
}: React.ComponentProps<"div"> & CarouselProps) {
	const [carouselRef, api] = useEmblaCarousel(
		{
			...opts,
			axis: orientation === "horizontal" ? "x" : "y",
		},
		plugins
	);
	const [canScrollPrev, setCanScrollPrev] = React.useState(false);
	const [canScrollNext, setCanScrollNext] = React.useState(false);

	const onSelect = React.useCallback((api: CarouselApi) => {
		if (!api) return;
		setCanScrollPrev(api.canScrollPrev());
		setCanScrollNext(api.canScrollNext());
	}, []);

	const scrollPrev = React.useCallback(() => {
		api?.scrollPrev();
	}, [api]);

	const scrollNext = React.useCallback(() => {
		api?.scrollNext();
	}, [api]);

	const handleKeyDown = React.useCallback(
		(event: React.KeyboardEvent<HTMLDivElement>) => {
			if (event.key === "ArrowLeft") {
				event.preventDefault();
				scrollPrev();
			} else if (event.key === "ArrowRight") {
				event.preventDefault();
				scrollNext();
			}
		},
		[scrollPrev, scrollNext]
	);

	React.useEffect(() => {
		if (!api || !setApi) return;
		setApi(api);
	}, [api, setApi]);

	React.useEffect(() => {
		if (!api) return;
		onSelect(api);
		api.on("reInit", onSelect);
		api.on("select", onSelect);

		return () => {
			api?.off("select", onSelect);
		};
	}, [api, onSelect]);

	return (
		<CarouselContext.Provider
			value={{
				carouselRef,
				api: api,
				opts,
				orientation: orientation || (opts?.axis === "y" ? "vertical" : "horizontal"),
				scrollPrev,
				scrollNext,
				canScrollPrev,
				canScrollNext,
				useDensityScaling,
			}}
		>
			<div
				onKeyDownCapture={handleKeyDown}
				className={cn("relative", className)}
				role="region"
				aria-roledescription="carousel"
				data-slot="carousel"
				{...props}
			>
				{children}
			</div>
		</CarouselContext.Provider>
	);
}

function CarouselContent({ className, ...props }: React.ComponentProps<"div">) {
	const { carouselRef, orientation } = useCarousel();

	return (
		<div ref={carouselRef} className="overflow-hidden" data-slot="carousel-content">
			<div
				className={cn("flex", orientation === "horizontal" ? "-ml-2 mb-6" : "-mt-4 flex-col", className)}
				{...props}
			/>
		</div>
	);
}

function CarouselItem({ className, ...props }: React.ComponentProps<"div">) {
	const { orientation, useDensityScaling } = useCarousel();
	const { densityStep } = useViewDensity();
	const [isHydrated, setIsHydrated] = React.useState(false);

	// Track hydration to prevent mismatch
	React.useEffect(() => {
		setIsHydrated(true);
	}, []);

	// Default basis classes (step 2 - lowest density)
	const defaultBasisClasses = [
		"basis-1/2",
		"sm:basis-1/2",
		"md:basis-1/3",
		"lg:basis-1/3",
		"xl:basis-1/4",
		"2xl:basis-1/5",
		"3xl:basis-1/6",
		"4xl:basis-1/8",
		"5xl:basis-1/10",
		"6xl:basis-1/12",
	];

	// Medium density basis classes (step 1)
	const mediumBasisClasses = [
		"basis-1/2",
		"sm:basis-1/2",
		"md:basis-1/4",
		"lg:basis-1/4",
		"xl:basis-1/5",
		"2xl:basis-1/6",
		"3xl:basis-1/7",
		"4xl:basis-1/9",
		"5xl:basis-1/11",
		"6xl:basis-1/13",
	];

	// High density basis classes (step 0 - highest density)
	const highBasisClasses = [
		"basis-1/2",
		"sm:basis-1/2",
		"md:basis-1/5",
		"lg:basis-1/5",
		"xl:basis-1/6",
		"2xl:basis-1/7",
		"3xl:basis-1/8",
		"4xl:basis-1/10",
		"5xl:basis-1/12",
		"6xl:basis-1/14",
	];

	// Select basis classes based on density step
	// Use medium density (step 1) as default until hydrated to match server render
	let basisClasses;
	if (useDensityScaling && isHydrated) {
		if (densityStep === 0) {
			basisClasses = highBasisClasses;
		} else if (densityStep === 1) {
			basisClasses = mediumBasisClasses;
		} else {
			basisClasses = defaultBasisClasses;
		}
	} else {
		// Use medium density classes for initial render (matches server default)
		basisClasses = mediumBasisClasses;
	}

	return (
		<div
			role="group"
			aria-roledescription="slide"
			data-slot="carousel-item"
			className={cn(
				"min-w-0 shrink-0 grow-0",
				...basisClasses,
				orientation === "horizontal" ? "pl-2 mb-0.5" : "pt-4",
				className
			)}
			{...props}
		/>
	);
}

function CarouselPrevious({
	className,
	variant = "outline",
	size = "icon",
	...props
}: React.ComponentProps<typeof Button>) {
	const { orientation, scrollPrev, canScrollPrev } = useCarousel();

	return (
		<Button
			data-slot="carousel-previous"
			variant={variant}
			size={size}
			className={cn(
				"absolute size-8 rounded-full",
				orientation === "horizontal"
					? "top-1/2 -left-12 -translate-y-1/2"
					: "-top-12 left-1/2 -translate-x-1/2 rotate-90",
				className
			)}
			disabled={!canScrollPrev}
			onClick={scrollPrev}
			{...props}
		>
			<ChevronLeft />
			<span className="sr-only">Previous slide</span>
		</Button>
	);
}

function CarouselNext({
	className,
	variant = "outline",
	size = "icon",
	...props
}: React.ComponentProps<typeof Button>) {
	const { orientation, scrollNext, canScrollNext } = useCarousel();

	return (
		<Button
			data-slot="carousel-next"
			variant={variant}
			size={size}
			className={cn(
				"absolute size-8 rounded-full",
				orientation === "horizontal"
					? "top-1/2 -right-12 -translate-y-1/2"
					: "-bottom-12 left-1/2 -translate-x-1/2 rotate-90",
				className
			)}
			disabled={!canScrollNext}
			onClick={scrollNext}
			{...props}
		>
			<ChevronRight />
			<span className="sr-only">Next slide</span>
		</Button>
	);
}

export { type CarouselApi, Carousel, CarouselContent, CarouselItem, CarouselPrevious, CarouselNext };
