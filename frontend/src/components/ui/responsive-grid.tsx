import React, { type PropsWithChildren } from "react";
import { cn } from "@/lib/utils"; // Assuming you have a cn utility
import Link from "next/link";

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
}

/**
 * A responsive grid component that adjusts the number of columns based on screen size.
 * The column counts are derived from a common basis pattern used throughout the application.
 * - Default: 2 columns
 * - sm: 2 columns
 * - md: 3 columns
 * - lg: 3 columns
 * - xl: 4 columns
 * - 2xl: 5 columns
 * - 3xl: 6 columns
 * - 4xl: 8 columns
 * - 5xl: 10 columns
 * - 6xl: 12 columns
 */
export const ResponsiveGrid: React.FC<
	PropsWithChildren<ResponsiveGridProps>
> = ({ className, title, link, maxItems, children }) => {
	const items = React.Children.toArray(children);
	const displayedItems = maxItems ? items.slice(0, maxItems) : items;

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
					// Default (2 columns)
					"grid-cols-2 [&>*:nth-child(n+5)]:hidden",
					// sm (2 columns)
					"sm:grid-cols-2 sm:[&>*:nth-child(n+5)]:block sm:[&>*:nth-child(n+7)]:hidden",
					// md (3 columns)
					"md:grid-cols-3 md:[&>*:nth-child(n+7)]:block md:[&>*:nth-child(n+10)]:hidden",
					// lg (3 columns)
					"lg:grid-cols-3 lg:[&>*:nth-child(n+10)]:block lg:[&>*:nth-child(n+13)]:hidden",
					// xl (4 columns)
					"xl:grid-cols-4 xl:[&>*:nth-child(n+13)]:block xl:[&>*:nth-child(n+17)]:hidden",
					// 2xl (5 columns)
					"2xl:grid-cols-5 2xl:[&>*:nth-child(n+17)]:block 2xl:[&>*:nth-child(n+22)]:hidden",
					// 3xl (6 columns)
					"3xl:grid-cols-6 3xl:[&>*:nth-child(n+22)]:block 3xl:[&>*:nth-child(n+28)]:hidden",
					// 4xl (8 columns)
					"4xl:grid-cols-8 4xl:[&>*:nth-child(n+28)]:block 4xl:[&>*:nth-child(n+36)]:hidden",
					// 5xl (10 columns)
					"5xl:grid-cols-10 5xl:[&>*:nth-child(n+36)]:block 5xl:[&>*:nth-child(n+46)]:hidden",
					// 6xl (12 columns)
					"6xl:grid-cols-12 6xl:[&>*:nth-child(n+46)]:block",
					className
				)}
			>
				{displayedItems}
			</div>
		</div>
	);
};
