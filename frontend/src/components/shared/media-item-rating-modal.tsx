"use client";

import { patchAddRatingToMediaItem } from "@/services/mediaserver/api-mediaserver-rate-media-item";
import { Star } from "lucide-react";

import { useEffect, useMemo, useState } from "react";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";

import { cn } from "@/lib/cn";

import type { MediaItem } from "@/types/media-and-posters/media-item-and-library";

export type MediaItemRatingModalProps = {
	mediaItem: MediaItem;
	isOpen: boolean;
	onClose: () => void;
};

const clampToHalfStar = (value: number) => {
	const clamped = Math.max(0, Math.min(5, value));
	return Math.round(clamped * 2) / 2;
};

const getInitialRatingFromMediaItem = (mediaItem: MediaItem) => {
	const userGuid = mediaItem.Guids?.find((g) => g.Provider === "user" && g.Rating);
	if (!userGuid?.Rating) return 0;

	const parsed = parseFloat(userGuid.Rating);
	return Number.isFinite(parsed) ? parsed : 0;
};

function StarVisual({ fill }: { fill: 0 | 0.5 | 1 }) {
	return (
		<span className="relative inline-flex h-6 w-6">
			<Star className="h-6 w-6 text-muted-foreground" strokeWidth={2} />
			{fill === 1 && (
				<Star className="absolute inset-0 h-6 w-6 text-yellow-500 fill-yellow-500" strokeWidth={2} />
			)}
			{fill === 0.5 && (
				<Star
					className={cn(
						"absolute inset-0 h-6 w-6 text-yellow-500 fill-yellow-500",
						"[clip-path:inset(0_50%_0_0)]"
					)}
					strokeWidth={2}
				/>
			)}
		</span>
	);
}

export function MediaItemRatingModal({ mediaItem, isOpen, onClose }: MediaItemRatingModalProps) {
	const [initialRating, setInitialRating] = useState(0);
	const [rating, setRating] = useState<number>(0);
	const [hoverRating, setHoverRating] = useState<number | null>(null);
	const [saving, setSaving] = useState(false);

	useEffect(() => {
		if (!isOpen) return;

		const start = clampToHalfStar(getInitialRatingFromMediaItem(mediaItem));
		setInitialRating(start);
		setRating(start);
		setHoverRating(null);
		setSaving(false);
	}, [isOpen, mediaItem]);

	const displayRating = hoverRating ?? rating;

	const starFills = useMemo(() => {
		return Array.from({ length: 5 }, (_, i) => {
			const starIndex = i + 1;
			const diff = displayRating - (starIndex - 1);
			if (diff >= 1) return 1 as const;
			if (diff >= 0.5) return 0.5 as const;
			return 0 as const;
		});
	}, [displayRating]);

	const handleSave = async () => {
		try {
			setSaving(true);
			const response = await patchAddRatingToMediaItem(mediaItem, rating);
			if (response.status === "error") {
				throw new Error(response.error?.message || "Unknown error rating media item");
			}

			onClose();
		} finally {
			setSaving(false);
		}
	};

	return (
		<Dialog
			open={isOpen}
			onOpenChange={(open) => {
				if (!open) onClose();
			}}
		>
			<DialogContent className={cn("sm:max-w-[520px]", "border border-primary")}>
				<DialogHeader>
					<DialogTitle className="text-lg font-bold">Rate</DialogTitle>
					<DialogDescription className="text-sm text-muted-foreground">
						Choose a rating for <span className="font-semibold">{mediaItem.Title}</span> ({mediaItem.Year})
					</DialogDescription>
				</DialogHeader>

				<div className="mt-2">
					<div className="flex items-center justify-between gap-3">
						<div
							className="flex items-center"
							onMouseLeave={() => setHoverRating(null)}
							aria-label="Rating"
						>
							{Array.from({ length: 5 }, (_, i) => {
								const starNumber = i + 1;

								return (
									<span key={starNumber} className="relative inline-flex">
										<button
											type="button"
											className="absolute inset-y-0 left-0 w-1/2 z-10"
											aria-label={`${starNumber - 0.5} stars`}
											onMouseEnter={() => setHoverRating(starNumber - 0.5)}
											onFocus={() => setHoverRating(starNumber - 0.5)}
											onClick={() => setRating(starNumber - 0.5)}
										/>
										<button
											type="button"
											className="absolute inset-y-0 right-0 w-1/2 z-10"
											aria-label={`${starNumber} stars`}
											onMouseEnter={() => setHoverRating(starNumber)}
											onFocus={() => setHoverRating(starNumber)}
											onClick={() => setRating(starNumber)}
										/>

										<span className="p-0.5">
											<StarVisual fill={starFills[i]} />
										</span>
									</span>
								);
							})}
						</div>

						<div className="text-sm text-muted-foreground tabular-nums">
							<span className="font-medium text-foreground">{displayRating.toFixed(1)}</span> / 5.0
						</div>
					</div>

					<div className="mt-3 flex items-center gap-2">
						<Button
							type="button"
							variant="outline"
							onClick={() => setRating(initialRating)}
							disabled={saving || rating === initialRating}
						>
							Reset
						</Button>

						<Button
							type="button"
							variant="outline"
							onClick={() => setRating((r) => clampToHalfStar(r - 0.5))}
							disabled={saving || rating <= 0}
						>
							-
						</Button>

						<Button
							type="button"
							variant="outline"
							onClick={() => setRating((r) => clampToHalfStar(r + 0.5))}
							disabled={saving || rating >= 5}
						>
							+
						</Button>
					</div>
				</div>

				<div className="mt-6 flex justify-end gap-2">
					<Button type="button" variant="outline" onClick={onClose} disabled={saving}>
						Cancel
					</Button>
					<Button type="button" onClick={handleSave} disabled={saving}>
						{saving ? "Saving..." : "Save"}
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
}
