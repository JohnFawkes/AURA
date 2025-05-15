"use client";

import { Button } from "@/components/ui/button";
import { ArrowUp } from "lucide-react";
import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";
import { usePathname } from "next/navigation";

export function JumpToTop() {
	const [isVisible, setIsVisible] = useState(false);
	const pathName = usePathname();

	const isRefreshPage = pathName === "/" || pathName === "/saved-sets/";
	const rightClass = isRefreshPage ? "right-15 sm:right-25" : "right-3";

	useEffect(() => {
		const toggleVisibility = () => {
			// Show button when page is scrolled up to given distance
			const scrollThreshold = 300;
			if (window.pageYOffset > scrollThreshold) {
				setIsVisible(true);
			} else {
				setIsVisible(false);
			}
		};

		window.addEventListener("scroll", toggleVisibility);

		return () => {
			window.removeEventListener("scroll", toggleVisibility);
		};
	}, []);

	const scrollToTop = () => {
		window.scrollTo({
			top: 0,
			behavior: "smooth",
		});
	};

	return (
		<Button
			variant="outline"
			size="sm"
			className={cn(
				`fixed z-100 ${rightClass} bottom-10 sm:bottom-15  rounded-full shadow-lg transition-all duration-300 bg-background border-primary-dynamic text-primary-dynamic hover:bg-primary-dynamic hover:text-primary cursor-pointer`,
				isVisible
					? "opacity-100 translate-y-0"
					: "opacity-0 translate-y-4 pointer-events-none"
			)}
			onClick={scrollToTop}
			aria-label="back to Top"
		>
			<ArrowUp className="h-3 w-3 mr-1" />
			<span className="text-xs hidden sm:inline">Back to Top</span>
		</Button>
	);
}
