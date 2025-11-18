"use client";

import React, { type PropsWithChildren } from "react";

import Link from "next/link";

import { useViewDensity } from "@/components/shared/view-density-context";

import { cn } from "@/lib/cn";

/**
 * Defines the size variants for the ResponsiveGrid component.
 */
type GridSizeVariant = "regular" | "larger";

/**
 * Defines the props for the ResponsiveGrid component.
 */
interface ResponsiveGridProps {
	/**
	 * Optional additional CSS class names to apply to the grid container.
	 */
	className?: string;

	/**
	 * Optional title to display above the grid.
	 */
	title?: string;

	/**
	 * Optional link URL for the title. If provided, the title becomes a link.
	 */
	link?: string;

	/**
	 * Maximum number of items to display at each breakpoint
	 * If not provided, all items will be rendered
	 */
	maxItems?: number;

	/**
	 * Whether to apply density scaling. Defaults to true.
	 */
	useDensityScaling?: boolean;

	/**
	 * Size variant of the grid. Affects the number of columns.
	 * - "regular": Default column count (more columns, smaller items)
	 * - "larger": Fewer columns for larger content items
	 * Defaults to "regular".
	 */
	size?: GridSizeVariant;
}

/**
 * A responsive grid component that adjusts the number of columns based on screen size.
 *
 * The number of columns can be dynamically adjusted based on user preference using density steps:
 * - Step 2: Default columns (lowest density)
 * - Step 1: +1 column (medium density)
 * - Step 0: +2 columns (highest density)
 */
export const ResponsiveGrid: React.FC<PropsWithChildren<ResponsiveGridProps>> = ({
	className,
	title,
	link,
	maxItems,
	useDensityScaling = true,
	size = "regular",
	children,
}) => {
	const items = React.Children.toArray(children);
	const displayedItems = maxItems ? items.slice(0, maxItems) : items;

	// Use density context for dynamic column adjustment
	const { densityStep } = useViewDensity();

	// Base classes for default density (step 2) - Regular size
	const baseRegularClasses = [
		"grid-cols-2",
		"sm:grid-cols-2",
		"md:grid-cols-3",
		"lg:grid-cols-4",
		"xl:grid-cols-5",
		"2xl:grid-cols-6",
		"3xl:grid-cols-7",
		"3.5xl:grid-cols-8",
		"4xl:grid-cols-9",
		"4.5xl:grid-cols-10",
		"5xl:grid-cols-12",
		"5.5xl:grid-cols-14",
		"6xl:grid-cols-16",
		"6.5xl:grid-cols-18",
		"7xl:grid-cols-20",
		"8xl:grid-cols-24",
	];

	// Medium density classes (step 1) for Regular size
	const mediumRegularClasses = [
		"grid-cols-2",
		"sm:grid-cols-3",
		"md:grid-cols-4",
		"lg:grid-cols-5",
		"xl:grid-cols-6",
		"2xl:grid-cols-8",
		"3xl:grid-cols-9",
		"3.5xl:grid-cols-10",
		"4xl:grid-cols-12",
		"4.5xl:grid-cols-14",
		"5xl:grid-cols-16",
		"5.5xl:grid-cols-18",
		"6xl:grid-cols-20",
		"6.5xl:grid-cols-22",
		"7xl:grid-cols-24",
		"8xl:grid-cols-28",
	];

	// Highest density classes (step 0) for Regular size
	const highRegularClasses = [
		"grid-cols-2",
		"sm:grid-cols-3",
		"md:grid-cols-5",
		"lg:grid-cols-6",
		"xl:grid-cols-8",
		"2xl:grid-cols-10",
		"3xl:grid-cols-12",
		"3.5xl:grid-cols-14",
		"4xl:grid-cols-16",
		"4.5xl:grid-cols-18",
		"5xl:grid-cols-20",
		"5.5xl:grid-cols-22",
		"6xl:grid-cols-24",
		"6.5xl:grid-cols-28",
		"7xl:grid-cols-32",
		"8xl:grid-cols-40",
	];

	// Base classes for default density (step 2) - Larger size
	const baseLargerClasses = [
		"grid-cols-2",
		"sm:grid-cols-2",
		"md:grid-cols-2",
		"lg:grid-cols-3",
		"xl:grid-cols-4",
		"2xl:grid-cols-5",
		"3xl:grid-cols-6",
		"3.5xl:grid-cols-7",
		"4xl:grid-cols-8",
		"4.5xl:grid-cols-9",
		"5xl:grid-cols-10",
		"5.5xl:grid-cols-12",
		"6xl:grid-cols-14",
		"6.5xl:grid-cols-16",
		"7xl:grid-cols-18",
		"8xl:grid-cols-20",
	];

	// Medium density classes (step 1) for Larger size
	const mediumLargerClasses = [
		"grid-cols-2",
		"sm:grid-cols-3",
		"md:grid-cols-3",
		"lg:grid-cols-4",
		"xl:grid-cols-5",
		"2xl:grid-cols-6",
		"3xl:grid-cols-7",
		"3.5xl:grid-cols-8",
		"4xl:grid-cols-9",
		"4.5xl:grid-cols-10",
		"5xl:grid-cols-12",
		"5.5xl:grid-cols-14",
		"6xl:grid-cols-16",
		"6.5xl:grid-cols-18",
		"7xl:grid-cols-20",
		"8xl:grid-cols-22",
	];

	// Highest density classes (step 0) for Larger size
	const highLargerClasses = [
		"grid-cols-2",
		"sm:grid-cols-3",
		"md:grid-cols-4",
		"lg:grid-cols-5",
		"xl:grid-cols-6",
		"2xl:grid-cols-8",
		"3xl:grid-cols-9",
		"3.5xl:grid-cols-10",
		"4xl:grid-cols-12",
		"4.5xl:grid-cols-14",
		"5xl:grid-cols-16",
		"5.5xl:grid-cols-18",
		"6xl:grid-cols-20",
		"6.5xl:grid-cols-22",
		"7xl:grid-cols-24",
		"8xl:grid-cols-28",
	];

	// Select classes based on size and density step
	let columnClasses;

	if (useDensityScaling) {
		if (size === "regular") {
			if (densityStep === 0) {
				columnClasses = highRegularClasses;
			} else if (densityStep === 1) {
				columnClasses = mediumRegularClasses;
			} else {
				columnClasses = baseRegularClasses;
			}
		} else {
			// "larger" size
			if (densityStep === 0) {
				columnClasses = highLargerClasses;
			} else if (densityStep === 1) {
				columnClasses = mediumLargerClasses;
			} else {
				columnClasses = baseLargerClasses;
			}
		}
	} else {
		// No density scaling - use base classes for the selected size
		columnClasses = size === "regular" ? baseRegularClasses : baseLargerClasses;
	}

	return (
		<div className="space-y-2">
			{title && (
				<h3 className="text-lg font-medium">
					{link ? (
						<Link href={link} className="hover:underline">
							{title}
						</Link>
					) : (
						title
					)}
				</h3>
			)}
			<div
				className={cn(
					"grid gap-1.5",
					...columnClasses,
					// Item visibility classes - modified to allow infinite scrolling
					// Only apply visibility restrictions when maxItems is provided
					...(maxItems
						? [
								"[&>*:nth-child(n+5)]:hidden",
								"sm:[&>*:nth-child(n+5)]:block sm:[&>*:nth-child(n+7)]:hidden",
								"md:[&>*:nth-child(n+7)]:block md:[&>*:nth-child(n+10)]:hidden",
								"lg:[&>*:nth-child(n+10)]:block lg:[&>*:nth-child(n+13)]:hidden",
								"xl:[&>*:nth-child(n+13)]:block xl:[&>*:nth-child(n+17)]:hidden",
								"2xl:[&>*:nth-child(n+17)]:block 2xl:[&>*:nth-child(n+22)]:hidden",
								"3xl:[&>*:nth-child(n+22)]:block 3xl:[&>*:nth-child(n+28)]:hidden",
								"4xl:[&>*:nth-child(n+28)]:block 4xl:[&>*:nth-child(n+36)]:hidden",
								"5xl:[&>*:nth-child(n+36)]:block 5xl:[&>*:nth-child(n+46)]:hidden",
								"6xl:[&>*:nth-child(n+46)]:block",
							]
						: []),
					className
				)}
			>
				{displayedItems}
			</div>
		</div>
	);
};
