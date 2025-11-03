"use client";

import { deleteFromQueue } from "@/services/download-queue/delete-item";
import { toast } from "sonner";

import React, { useState } from "react";

import Link from "next/link";

import { AssetImage } from "@/components/shared/asset-image";
import { ConfirmDestructiveDialogActionButton } from "@/components/shared/dialog-destructive-action";
import DownloadModal from "@/components/shared/download-modal";
import { renderTypeBadges } from "@/components/shared/saved-sets-shared";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { H4, P } from "@/components/ui/typography";

import { useMediaStore } from "@/lib/stores/global-store-media-store";

import { DBMediaItemWithPosterSets } from "@/types/database/db-poster-set";

const DownloadQueueEntry: React.FC<{
	entry: DBMediaItemWithPosterSets;
	fetchQueueEntries?: () => Promise<void>;
}> = ({ entry, fetchQueueEntries }) => {
	// Initialize edit state from the savedSet.PosterSets array.
	const [editSets] = useState(() =>
		entry.PosterSets.map((set) => ({
			id: set.PosterSetID,
			previousDateUpdated: set.PosterSet.DateUpdated,
			set: set.PosterSet || "Unknown",
			selectedTypes: set.SelectedTypes,
			autoDownload: set.AutoDownload,
			toDelete: false,
		}))
	);

	// Access global stores
	const { setMediaItem } = useMediaStore();

	const onDeleteConfirm = async () => {
		try {
			const response = await deleteFromQueue(entry);
			if (response.status === "error") {
				toast.error(
					`Error deleting from queue: ${response.error?.message || "Unknown error occurred trying to delete."}`
				);
			} else {
				toast.success(response.data || "Successfully deleted from queue.");
			}
		} catch (error) {
			toast.error(
				`Error deleting from queue: ${error instanceof Error ? error.message : "Unknown error occurred trying to delete."}`
			);
		}
		if (fetchQueueEntries) {
			await fetchQueueEntries();
		}
	};

	return (
		<Card className="relative w-full max-w-md mx-auto mb-4">
			<CardHeader>
				{/* Top Left: Delete File */}
				<div className="absolute top-2 left-2">
					<ConfirmDestructiveDialogActionButton
						variant="ghost"
						className="text-destructive border-1 shadow-none hover:text-red-500 cursor-pointer"
						confirmText="Delete File"
						title="Delete Downloaded File?"
						description="Are you sure you want to delete the downloaded file for this media item? This action cannot be undone."
						onConfirm={onDeleteConfirm}
					>
						Delete
					</ConfirmDestructiveDialogActionButton>
				</div>
				{/* Top Right: Dropdown Menu */}
				<div className="absolute top-2 right-2 justify-end">
					<DownloadModal
						setID={editSets[0].id}
						setTitle={editSets[0].set.Title}
						setType={editSets[0].set.Type}
						setAuthor={editSets[0].set.User.Name}
						posterSets={[editSets[0].set]}
						autoDownloadDefault={editSets[0].autoDownload}
					/>
				</div>
				{/* Middle: Image */}
				<div className="flex justify-center mt-6">
					<AssetImage
						image={entry.MediaItem}
						className="w-[170px] h-auto transition-transform hover:scale-105"
					/>
				</div>
			</CardHeader>

			{/* Content */}
			<CardContent>
				{/* Title */}
				<H4>
					<Link
						//href={formatMediaItemUrl(entry.MediaItem)}
						href={"/media-item/"}
						className="text-primary hover:underline"
						onClick={() => {
							setMediaItem(entry.MediaItem);
						}}
					>
						{entry.MediaItem.Title}
					</Link>
				</H4>

				{/* Year */}
				<P className="text-sm text-muted-foreground">Year: {entry.MediaItem.Year}</P>

				{/* Library Title */}
				<P className="text-sm text-muted-foreground">Library: {entry.MediaItem.LibraryTitle}</P>

				<Separator className="my-4" />

				{entry.PosterSets.some(
					(set) => Array.isArray(set.SelectedTypes) && set.SelectedTypes.some((type) => type.trim() !== "")
				) ? (
					<div className="flex flex-wrap gap-2">{renderTypeBadges(entry)}</div>
				) : (
					<div className="flex flex-wrap gap-2">
						<Badge key={"no-types"} variant="outline" className="text-sm bg-red-500">
							No Selected Types
						</Badge>
					</div>
				)}
			</CardContent>
		</Card>
	);
};

export default DownloadQueueEntry;
