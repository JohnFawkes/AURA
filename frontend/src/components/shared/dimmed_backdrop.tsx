import { useEffect, useState } from "react";

import Image from "next/image";

import { cn } from "@/lib/cn";

export function DimmedBackground({ backdropURL }: { backdropURL: string }) {
	const [isBlurred, setIsBlurred] = useState(false);
	const [imageValid, setImageValid] = useState(true);

	useEffect(() => {
		if (!backdropURL) {
			setImageValid(false);
			return;
		}
		const img = new window.Image();
		img.src = backdropURL;
		img.onload = () => setImageValid(true);
		img.onerror = () => setImageValid(false);
	}, [backdropURL]);

	useEffect(() => {
		const handleScroll = () => {
			if (window.scrollY > 300) {
				setIsBlurred(true);
			} else {
				setIsBlurred(false);
			}
		};
		window.addEventListener("scroll", handleScroll);
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
					{imageValid ? (
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
					) : (
						<div
							className="absolute inset-0 animate-pulse bg-gradient-to-r from-red-800 via-black-700 to-grey-800"
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
					)}
				</div>
			</div>
		</div>
	);
}
