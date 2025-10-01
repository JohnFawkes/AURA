import { useEffect, useState } from "react";

import Image from "next/image";

import { cn } from "@/lib/cn";

export function DimmedBackground({ backdropURL }: { backdropURL: string }) {
	const [isBlurred, setIsBlurred] = useState(false);

	// Handle scroll event to blur the background
	useEffect(() => {
		const handleScroll = () => {
			// Check if the user has scrolled down 300px (adjust as needed)
			if (window.scrollY > 300) {
				setIsBlurred(true);
			} else {
				setIsBlurred(false);
			}
		};

		// Add scroll event listener
		window.addEventListener("scroll", handleScroll);

		// Cleanup event listener on component unmount
		return () => {
			window.removeEventListener("scroll", handleScroll);
		};
	}, []);

	return (
		<div
			className={cn(
				"fixed inset-0 -z-20 overflow-hidden h-full transition-[filter] duration-1000",
				isBlurred && "blur-md"
			)}
			style={{ width: "100vw" }}
		>
			<div className="absolute inset-0 bg-background">
				<div className="absolute inset-0 opacity-[0.015] mix-blend-overlay">
					<div className="absolute inset-0 bg-[url(/gradient.svg)]"></div>{" "}
				</div>

				<div
					className="absolute inset-0"
					style={
						{
							background: `
                    radial-gradient(ellipse at 30% 30%, var(--dynamic-left) 0%, transparent 60%),
                    radial-gradient(ellipse at bottom right, var(--dynamic-bottom) 0%, transparent 60%),
                    radial-gradient(ellipse at center, var(--dynamic-dark-muted) 0%, transparent 80%),
                    var(--background)
                  `,
							opacity: 0.5,
						} as React.CSSProperties
					}
				/>
			</div>

			<div className="absolute top-15 right-0 w-full lg:w-[70vw] aspect-[16/9] z-50">
				<div className="relative w-full h-full">
					<Image
						src={backdropURL}
						alt={"Backdrop"}
						fill
						unoptimized
						loading="lazy"
						className="object-cover object-right-top"
						style={{
							maskImage: `url(/gradient.svg)`,
							WebkitMaskImage: `url(/gradient.svg)`,
							maskSize: "100% 100%",
							WebkitMaskSize: "100% 100%",
							maskRepeat: "no-repeat",
							WebkitMaskRepeat: "no-repeat",
							maskPosition: "center",
							WebkitMaskPosition: "center",
						}}
					/>
				</div>
			</div>
		</div>
	);
}
