import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import { useEffect, useState } from "react";
import { useInView } from "react-intersection-observer";
import { fetchMediaServerImageData } from "../services/api.mediaserver";

const imageCache = new Map<string, string>(); // Cache to store loaded image URLs

const PlexPosterImage: React.FC<{ ratingKey: string; alt: string }> = ({
	ratingKey,
	alt,
}) => {
	const [imageSrc, setImageSrc] = useState<string | null>(null);
	const { ref, inView } = useInView({ triggerOnce: true });

	useEffect(() => {
		// Reset the image source when the ratingKey changes
		setImageSrc(imageCache.get(ratingKey) || null);

		if (inView && !imageCache.has(ratingKey)) {
			const loadImage = async () => {
				try {
					const posterURL = await fetchMediaServerImageData(
						ratingKey,
						"poster"
					);
					if (!posterURL) {
						throw new Error("Image not found");
					}
					imageCache.set(ratingKey, posterURL); // Cache the image URL
					setImageSrc(posterURL);
				} catch {
					setImageSrc("/logo.png"); // Fallback image
				}
			};
			loadImage();
		}
	}, [inView, ratingKey]);

	return (
		<Box
			ref={ref}
			sx={{
				width: 100,
				height: 160,
				flexShrink: 0,
				borderRadius: 1,
				overflow: "hidden",
			}}
		>
			{imageSrc ? (
				<img
					src={imageSrc}
					alt={alt}
					style={{
						width: "100%",
						height: "100%",
						objectFit: "cover",
					}}
				/>
			) : (
				<Box
					sx={{
						width: "100%",
						height: "100%",
						display: "flex",
						justifyContent: "center",
						alignItems: "center",
					}}
				>
					<Typography variant="caption" color="text.secondary">
						Loading...
					</Typography>
				</Box>
			)}
		</Box>
	);
};

export default PlexPosterImage;
