"use client";

import { useEffect, useState } from "react";

import Image from "next/image";

import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

// Responsive skeleton count hook
function useSkeletonCount() {
	const [count, setCount] = useState(3);

	useEffect(() => {
		function handleResize() {
			if (window.innerWidth >= 1024) setCount(3);
			else if (window.innerWidth >= 768) setCount(2);
			else setCount(1);
		}
		handleResize();
		window.addEventListener("resize", handleResize);
		return () => window.removeEventListener("resize", handleResize);
	}, []);

	return count;
}

// Single skeleton card
export const HomeMediaItemCardSkeleton: React.FC = () => (
	<Card className="relative items-center">
		<div className="absolute top-2 left-2 z-10">
			<Skeleton className="h-5 w-5 rounded-full bg-green-300" />
		</div>
		<div className="relative w-[170px] h-[255px] mx-auto mb-2 flex items-center justify-center">
			<Skeleton className="absolute inset-0 w-full h-full rounded-lg" />
			<Image
				src="/aura_logo.svg"
				alt="Aura Logo"
				width={48}
				height={48}
				className="opacity-70 animate-pulse"
				draggable={false}
			/>
		</div>
		<Skeleton className="h-6 w-[80%] mx-auto mb-2" />
		<CardContent className="flex justify-center gap-2">
			<Skeleton className="h-5 w-12 rounded" />
			<Skeleton className="h-5 w-20 rounded" />
		</CardContent>
	</Card>
);

// Grid of skeleton cards, centered and responsive
export const HomeMediaItemCardSkeletonGrid: React.FC = () => {
	const skeletonCount = useSkeletonCount();

	return (
		<div className="flex justify-center items-center w-full">
			<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mx-auto max-w-screen-xl">
				{Array.from({ length: skeletonCount }).map((_, idx) => (
					<HomeMediaItemCardSkeleton key={idx} />
				))}
			</div>
		</div>
	);
};
