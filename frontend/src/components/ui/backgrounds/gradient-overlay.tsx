"use client";

import { cn } from "@/lib/utils";

interface GradientOverlayProps {
	className?: string;
}

/**
 * Renders a gradient overlay using CSS variables set by DynamicColour component
 *
 * CSS variables used:
 * --overlay-tl: Top-left color
 * --overlay-ml: Middle-left color
 * --overlay-bl: Bottom-left color
 * --overlay-bm: Bottom-middle color
 */
export function GradientOverlay({ className }: GradientOverlayProps) {
	return (
		<div className="w-full h-full -z-10">
			<div
				className={cn(
					"fixed inset-0 -z-20 w-full h-full bg-radial-[at_70%_0%] from-transparent from-30% via-overlay-tl via-40% to-overlay-tl to-60%",
					className
				)}
			/>
			<div
				className={cn(
					"fixed inset-0 -z-20 w-full h-full bg-radial-[at_80%_10%] from-transparent from-25% via-overlay-tl via-35% to-overlay-tl to-50%",
					className
				)}
			/>
			<div
				className={cn(
					"fixed inset-0 -z-20 w-full h-full bg-radial-[at_90%_75%] from-overlay-bl/40 from-0% via-overlay-bl/40 via-35% to-transparent to-50%",
					className
				)}
			/>
		</div>
	);
}
