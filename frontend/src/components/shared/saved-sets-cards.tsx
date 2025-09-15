"use client";

import { formatMediaItemUrl } from "@/helper/formatMediaItemURL";
import { Delete, Edit, MoreHorizontal, RefreshCcw, RefreshCwOff } from "lucide-react";

import React, { useState } from "react";

import Link from "next/link";

import { AssetImage } from "@/components/shared/asset-image";
import Loader from "@/components/shared/loader";
import {
	SavedSetDeleteModal,
	SavedSetEditModal,
	SavedSetsList,
	onCloseSavedSetsEditDeleteModals,
	refreshPosterSet,
	renderTypeBadges,
	savedSetsConfirmDelete,
	savedSetsConfirmEdit,
} from "@/components/shared/saved-sets-shared";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Separator } from "@/components/ui/separator";
import { H4, P } from "@/components/ui/typography";

import { useMediaStore } from "@/lib/mediaStore";
import { usePosterSetsStore } from "@/lib/posterSetStore";

import { APIResponse } from "@/types/apiResponse";
import { DBMediaItemWithPosterSets } from "@/types/databaseSavedSet";

const SavedSetsCard: React.FC<{
	savedSet: DBMediaItemWithPosterSets;
	onUpdate: () => void;
	handleRecheckItem: (title: string, item: DBMediaItemWithPosterSets) => void;
}> = ({ savedSet, onUpdate, handleRecheckItem }) => {
	// Initialize edit state from the savedSet.PosterSets array.
	const [editSets, setEditSets] = useState(() =>
		savedSet.PosterSets.map((set) => ({
			id: set.PosterSetID,
			previousDateUpdated: set.PosterSet.DateUpdated,
			set: set.PosterSet || "Unknown",
			selectedTypes: set.SelectedTypes,
			autoDownload: set.AutoDownload,
			toDelete: false,
		}))
	);

	// State to track Modal visibility
	const [isEditModalOpen, setIsEditModalOpen] = useState(false);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);

	// State to track any error messages during updates.
	const [updateError, setUpdateError] = useState<APIResponse<unknown> | null>(null);

	// State to prevent multiple simultaneous operations.
	const [isMounted, setIsMounted] = useState(false);

	// Access global stores
	const { setPosterSets, setSetAuthor, setSetID, setSetTitle, setSetType } = usePosterSetsStore();
	const { setMediaItem } = useMediaStore();

	// State to track if we are currently refreshing poster sets
	const [isRefreshing, setIsRefreshing] = useState(false);

	// Check if all sets are marked for deletion
	const allToDelete = editSets.every((set) => set.toDelete);

	// Track unignore loading state
	const [unignoreLoading, setUnignoreLoading] = useState(false);
	const onlyIgnore = savedSet.PosterSets.length === 1 && savedSet.PosterSets[0].PosterSetID === "ignore";

	return (
		<Card className="relative w-full max-w-md mx-auto mb-4">
			<CardHeader>
				{isRefreshing && (
					<div className="absolute inset-0 flex items-center justify-center bg-black/50 z-10">
						<Loader className="animate-spin h-8 w-8 text-primary" />
						<span className="ml-2 text-white">Refreshing sets...</span>
					</div>
				)}

				{/* Top Left: Auto Download Icon */}
				{savedSet.MediaItem.Type === "show" && (
					<div className="absolute top-2 left-2">
						{savedSet.PosterSets.some((set) => set.AutoDownload) ? (
							<RefreshCcw className="text-green-500" size={24} />
						) : (
							<RefreshCwOff className="text-red-500" size={24} />
						)}
					</div>
				)}
				{/* Top Right: Dropdown Menu */}
				<div className="absolute top-2 right-2">
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button
								variant="ghost"
								className="cursor-pointer p-1 hover:bg-muted/50 focus:bg-muted/50"
								size="icon"
							>
								<MoreHorizontal />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							<DropdownMenuItem
								className="cursor-pointer"
								onClick={async () => {
									await refreshPosterSet({
										editSets,
										setEditSets,
										savedSet,
										setIsRefreshing,
										setUpdateError,
									});
									setIsEditModalOpen(true);
								}}
							>
								<Edit className="ml-2" />
								{isRefreshing ? "Refreshing..." : "Edit"}
							</DropdownMenuItem>

							{savedSet.PosterSets.some((set) => set.AutoDownload) && (
								<DropdownMenuItem
									className="cursor-pointer"
									onClick={() => {
										handleRecheckItem(savedSet.MediaItem.Title, savedSet);
									}}
								>
									<RefreshCcw className="ml-2" />
									Force Autodownload Recheck
								</DropdownMenuItem>
							)}
							<DropdownMenuItem
								onClick={() => setIsDeleteModalOpen(true)}
								className="text-destructive cursor-pointer"
							>
								<Delete className="ml-2" />
								Delete
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				</div>
				{/* Middle: Image */}
				<div className="flex justify-center mt-6">
					<AssetImage
						image={savedSet.MediaItem}
						className="w-[170px] h-auto transition-transform hover:scale-105"
					/>
				</div>
			</CardHeader>

			{/* Content */}
			<CardContent>
				{/* Title */}
				<H4>
					<Link
						href={formatMediaItemUrl(savedSet.MediaItem)}
						className="text-primary hover:underline"
						onClick={() => {
							setMediaItem(savedSet.MediaItem);
						}}
					>
						{savedSet.MediaItem.Title}
					</Link>
				</H4>

				{/* Year */}
				<P className="text-sm text-muted-foreground">Year: {savedSet.MediaItem.Year}</P>

				{/* Library Title */}
				<P className="text-sm text-muted-foreground">Library: {savedSet.MediaItem.LibraryTitle}</P>

				{/* Last Updated */}
				<P className="text-sm text-muted-foreground">
					Last Updated:{" "}
					{(() => {
						const latestTimestamp = Math.max(
							...savedSet.PosterSets.map((ps) => new Date(ps.LastDownloaded).getTime())
						);
						const latestDate = new Date(latestTimestamp);
						return `${latestDate.toLocaleDateString("en-US")} at ${latestDate.toLocaleTimeString("en-US", {
							hour: "numeric",
							minute: "numeric",
							second: "numeric",
							hour12: true,
						})}`;
					})()}
				</P>

				<SavedSetsList
					savedSet={savedSet}
					layout="table"
					onUpdate={onUpdate}
					unignoreLoading={unignoreLoading}
					setUnignoreLoading={setUnignoreLoading}
					setUpdateError={setUpdateError}
					onSelectSet={(ps) => {
						setPosterSets([ps.PosterSet]);
						setSetType(ps.PosterSet.Type);
						setSetTitle(ps.PosterSet.Title);
						setSetAuthor(ps.PosterSet.User.Name);
						setSetID(ps.PosterSetID);
					}}
				/>
				<Separator className="my-4" />

				{savedSet.PosterSets.some(
					(set) => Array.isArray(set.SelectedTypes) && set.SelectedTypes.some((type) => type.trim() !== "")
				) ? (
					<div className="flex flex-wrap gap-2">{renderTypeBadges(savedSet)}</div>
				) : (
					<P className="text-sm text-muted-foreground">No types selected.</P>
				)}
			</CardContent>

			{/* Edit Modal */}
			<SavedSetEditModal
				open={isEditModalOpen}
				onClose={() =>
					onCloseSavedSetsEditDeleteModals({
						setIsEditModalOpen,
						setIsDeleteModalOpen,
						setUpdateError,
						setIsMounted,
					})
				}
				editSets={editSets}
				setEditSets={setEditSets}
				savedSet={savedSet}
				onlyIgnore={onlyIgnore}
				allToDelete={allToDelete}
				updateError={updateError}
				confirmEdit={() =>
					savedSetsConfirmEdit({
						editSets,
						savedSet,
						onUpdate,
						isMounted,
						setIsMounted,
						setUpdateError,
						setIsEditModalOpen,
						setIsDeleteModalOpen,
						allToDelete,
					})
				}
			/>

			{/* Delete Modal */}
			<SavedSetDeleteModal
				open={isDeleteModalOpen}
				onClose={() =>
					onCloseSavedSetsEditDeleteModals({
						setIsEditModalOpen,
						setIsDeleteModalOpen,
						setUpdateError,
						setIsMounted,
					})
				}
				title={savedSet.MediaItem.Title}
				confirmDelete={() =>
					savedSetsConfirmDelete({
						savedSet,
						onUpdate,
						isMounted,
						setIsMounted,
						setUpdateError,
						setIsDeleteModalOpen,
					})
				}
			/>
		</Card>
	);
};

export default SavedSetsCard;
