"use client";
import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardContent } from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogFooter,
	DialogTitle,
} from "@/components/ui/dialog";
import { H4, P, Small } from "@/components/ui/typography";
import { deleteMediaItemFromDB, patchSavedItemInDB } from "@/services/api.db";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "./dropdown-menu";
import { Badge } from "./badge";
import { Separator } from "./separator";
import { CheckCircle2 as Checkmark, X, MoreHorizontal } from "lucide-react";
import Image from "next/image";
import { DialogDescription } from "@radix-ui/react-dialog";
import Link from "next/link";
import { DBMediaItemWithPosterSets } from "@/types/databaseSavedSet";

const SavedSetsCard: React.FC<{
	savedSet: DBMediaItemWithPosterSets;
	onUpdate: () => void;
}> = ({ savedSet, onUpdate }) => {
	// Initialize edit state from the savedSet.PosterSets array.
	const [editSets, setEditSets] = useState(() =>
		savedSet.PosterSets.map((set) => ({
			id: set.PosterSetID,
			set: set.PosterSet || "Unknown",
			selectedTypes: set.SelectedTypes,
			autoDownload: set.AutoDownload,
			toDelete: false,
		}))
	);

	const [isEditModalOpen, setIsEditModalOpen] = useState(false);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [updateError, setUpdateError] = useState("");
	const [isMounted, setIsMounted] = useState(false);

	const allToDelete = editSets.every((set) => set.toDelete);

	const onClose = () => {
		setIsEditModalOpen(false);
		setIsDeleteModalOpen(false);
		setUpdateError("");
		setIsMounted(false);
	};

	const confirmEdit = async () => {
		if (isMounted) return;
		setIsMounted(true);

		if (allToDelete) {
			setIsEditModalOpen(false);
			setIsDeleteModalOpen(true);
			setUpdateError("");
			setIsMounted(false);
			return;
		}

		// Create a new DBSavedItem object with updated values
		const updatedSavedSet: DBMediaItemWithPosterSets = {
			...savedSet,
			PosterSets: editSets.map((editSet, idx) => ({
				PosterSetID: editSet.id,
				PosterSet: editSet.set,
				PosterSetJSON:
					typeof editSet.set === "object"
						? JSON.stringify(editSet.set)
						: savedSet.PosterSets[idx]?.PosterSetJSON || "",
				LastDownloaded: new Date().toISOString(),
				SelectedTypes: editSet.selectedTypes,
				AutoDownload: editSet.autoDownload,
				ToDelete: editSet.toDelete,
			})),
		};

		const response = await patchSavedItemInDB(updatedSavedSet);

		if (response.status !== "success") {
			setUpdateError(response.message);
		} else {
			setUpdateError("");
			setIsEditModalOpen(false);
			onUpdate();
		}

		setIsMounted(false);
	};

	const confirmDelete = async () => {
		if (isMounted) return;
		setIsMounted(true);
		const resp = await deleteMediaItemFromDB(savedSet.MediaItemID);
		if (resp.status !== "success") {
			setUpdateError(resp.message);
		} else {
			setIsDeleteModalOpen(false);
			setUpdateError("");
			onUpdate();
		}
		setIsMounted(false);
	};

	const renderSetBadges = () => {
		return savedSet.PosterSets.map((set) => (
			<Link
				key={set.PosterSetID}
				href={`https://mediux.pro/sets/${set.PosterSetID}`}
				target="_blank"
				rel="noopener noreferrer"
				className="transition transform hover:scale-105 hover:underline"
			>
				<Badge className="cursor-pointer text-sm">
					{set.PosterSetID}
				</Badge>
			</Link>
		));
	};

	// Replace the hard-coded array with dynamically generated list.
	const renderEditTypeBadges = (
		editSet: (typeof editSets)[number],
		index: number
	) => {
		const availableTypes: string[] = [];
		if (editSet.set && editSet.set.Poster) {
			availableTypes.push("poster");
		}
		if (editSet.set && editSet.set.Backdrop) {
			availableTypes.push("backdrop");
		}
		if (
			editSet.set &&
			editSet.set.SeasonPosters &&
			editSet.set.SeasonPosters.length > 0
		) {
			// Check to see if any of the Season Posters are Season 0
			const hasSeason0 = editSet.set.SeasonPosters.some(
				(season) => season.Season?.Number === 0
			);
			if (hasSeason0) {
				availableTypes.push("specialSeasonPoster");
			}
			// Check to see if any of the Season Posters are not Season 0
			const hasNonSeason0 = editSet.set.SeasonPosters.some(
				(season) => season.Season?.Number !== 0
			);
			if (hasNonSeason0) {
				availableTypes.push("seasonPoster");
			}
		}
		if (
			editSet.set &&
			editSet.set.TitleCards &&
			editSet.set.TitleCards.length > 0
		) {
			availableTypes.push("titlecard");
		}

		return availableTypes.map((type) => {
			const isSelected = editSet.selectedTypes.includes(type);
			// Disable the type if this set is marked for deletion
			// or if it's not selected here, but found in any other set.
			const isTypeDisabled =
				editSet.toDelete ||
				(!isSelected &&
					editSets.some(
						(item, j) =>
							j !== index && item.selectedTypes.includes(type)
					));
			return (
				<Badge
					key={type}
					className={`flex items-center gap-2 transition duration-200 ${
						isTypeDisabled
							? "bg-secondary opacity-50 cursor-not-allowed"
							: isSelected
							? "cursor-pointer bg-primary text-primary-foreground hover:bg-red-500"
							: "cursor-pointer bg-secondary text-secondary-foreground"
					}`}
					onClick={() => {
						if (isTypeDisabled) return;
						setEditSets((prev) =>
							prev.map((item, i) => {
								if (i !== index) return item;
								const newSelectedTypes =
									item.selectedTypes.includes(type)
										? item.selectedTypes.filter(
												(t) => t !== type
										  )
										: [...item.selectedTypes, type];
								return {
									...item,
									selectedTypes: newSelectedTypes,
								};
							})
						);
					}}
				>
					{type === "poster"
						? "Poster"
						: type === "backdrop"
						? "Backdrop"
						: type === "seasonPoster"
						? "Season Posters"
						: type === "specialSeasonPoster"
						? "Special Poster"
						: type === "titlecard"
						? "Title Card"
						: type}
				</Badge>
			);
		});
	};

	const renderTypeBadges = () => {
		// Flatten all SelectedTypes arrays from every poster set.
		const allTypes = savedSet.PosterSets.flatMap(
			(set) => set.SelectedTypes
		);
		const uniqueTypes = Array.from(new Set(allTypes));
		return uniqueTypes.map((type) =>
			// Check if the type is empty or not
			type.trim() === "" ? null : (
				// Render the badge only if the type is not empty
				<Badge key={type}>
					{type === "poster"
						? "Poster"
						: type === "backdrop"
						? "Backdrop"
						: type === "seasonPoster"
						? "Season Posters"
						: type === "specialSeasonPoster"
						? "Special Poster"
						: type === "titlecard"
						? "Title Card"
						: type}
				</Badge>
			)
		);
	};

	return (
		<Card className="relative w-full max-w-md mx-auto mb-4">
			<CardHeader>
				{/* Top Left: Auto Download Icon */}
				{savedSet.MediaItem.Type === "show" && (
					<div className="absolute top-2 left-2">
						{savedSet.PosterSets[0].AutoDownload ? (
							<Checkmark className="text-green-500" size={24} />
						) : (
							<X className="text-red-500" size={24} />
						)}
					</div>
				)}

				{/* Top Right: Dropdown Menu */}
				<div className="absolute top-2 right-2">
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button variant="ghost" size="icon">
								<MoreHorizontal />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							<DropdownMenuItem
								onClick={() => setIsEditModalOpen(true)}
							>
								Edit
							</DropdownMenuItem>
							<DropdownMenuItem
								onClick={() => setIsDeleteModalOpen(true)}
								className="text-destructive"
							>
								Delete
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				</div>

				{/* Middle: Image */}
				<div className="flex justify-center mt-6">
					<Image
						src={`/api/mediaserver/image/${savedSet.MediaItem.RatingKey}/poster`}
						alt={savedSet.MediaItem.Title}
						width={150}
						height={225}
						className="rounded-md"
						unoptimized
						loading="lazy"
					/>
				</div>
			</CardHeader>

			{/* Content */}
			<CardContent>
				{/* Title */}
				<H4 className="text-lg font-semibold">
					{savedSet.MediaItem.Title}
				</H4>

				{/* Year */}
				<P className="text-sm text-muted-foreground">
					Year: {savedSet.MediaItem.Year}
				</P>

				{/* Library Title */}
				<P className="text-sm text-muted-foreground">
					Library: {savedSet.MediaItem.LibraryTitle}
				</P>

				{/* Last Updated */}
				<P className="text-sm text-muted-foreground">
					Last Updated:{" "}
					{(() => {
						const latestTimestamp = Math.max(
							...savedSet.PosterSets.map((ps) =>
								new Date(ps.LastDownloaded).getTime()
							)
						);
						const latestDate = new Date(latestTimestamp);
						return `${latestDate.toLocaleDateString(
							"en-US"
						)} at ${latestDate.toLocaleTimeString("en-US", {
							hour: "numeric",
							minute: "numeric",
							second: "numeric",
							hour12: true,
						})}`;
					})()}
				</P>

				<div className="flex flex-wrap gap-2">
					{savedSet.PosterSets.length > 1 ? "Sets:" : "Set:"}
					{renderSetBadges()}
				</div>

				<Separator className="my-4" />

				{savedSet.PosterSets.some(
					(set) =>
						Array.isArray(set.SelectedTypes) &&
						set.SelectedTypes.some((type) => type.trim() !== "")
				) ? (
					<div className="flex flex-wrap gap-2">
						{renderTypeBadges()}
					</div>
				) : (
					<P className="text-sm text-muted-foreground">
						No types selected.
					</P>
				)}
			</CardContent>

			{/* Edit Modal */}
			<Dialog open={isEditModalOpen} onOpenChange={onClose}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Edit Saved Set</DialogTitle>
						<DialogDescription>
							Edit each set individually. Toggle type badges to
							update selected types. Use the delete option to mark
							a set for deletion.
						</DialogDescription>
					</DialogHeader>
					<div className="space-y-4">
						{editSets.map((editSet, index) => (
							<div
								key={editSet.id}
								className="border p-2 rounded-md"
							>
								<div className="flex items-center justify-between">
									<span className="font-semibold">
										Set ID:{" "}
										<Link
											href={`https://mediux.pro/sets/${editSet.id}`}
											target="_blank"
											rel="noopener noreferrer"
											className="hover:underline"
										>
											{editSet.id}
										</Link>
									</span>
									<Button
										variant={
											editSet.toDelete
												? "destructive"
												: "outline"
										}
										size="sm"
										onClick={() => {
											setEditSets((prev) =>
												prev.map((item, i) =>
													i === index
														? {
																...item,
																toDelete:
																	!item.toDelete,
																// Clear the selected types when marking for deletion.
																selectedTypes:
																	!item.toDelete
																		? []
																		: item.selectedTypes,
														  }
														: item
												)
											);
										}}
									>
										{editSet.toDelete
											? "Undo Delete"
											: "Delete Set"}
									</Button>
								</div>
								<div className="flex flex-wrap gap-2 mt-2">
									{renderEditTypeBadges(editSet, index)}
								</div>
								{savedSet.MediaItem.Type === "show" && (
									<div className="flex flex-wrap gap-2 mt-2">
										<Badge
											className={`cursor-pointer transition duration-200 ${
												editSet.autoDownload
													? "bg-primary text-primary-foreground hover:bg-red-500"
													: "bg-secondary text-secondary-foreground"
											}`}
											onClick={() => {
												setEditSets((prev) =>
													prev.map((item, i) =>
														i === index
															? {
																	...item,
																	autoDownload:
																		!item.autoDownload,
															  }
															: item
													)
												);
											}}
										>
											{editSet.autoDownload
												? "Autodownload"
												: "No Autodownload"}
										</Badge>
									</div>
								)}
							</div>
						))}
					</div>
					{updateError && (
						<Small className="text-destructive mt-2">
							{updateError}
						</Small>
					)}
					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => {
								setEditSets(
									savedSet.PosterSets.map((set) => ({
										id: set.PosterSetID,
										set: set.PosterSet,
										selectedTypes: set.SelectedTypes,
										autoDownload: set.AutoDownload,
										toDelete: false,
									}))
								);
								onClose();
							}}
						>
							Cancel
						</Button>
						<Button
							variant={allToDelete ? "destructive" : "default"}
							onClick={confirmEdit}
						>
							{allToDelete
								? editSets.length === 1
									? "Delete Set"
									: "Delete All"
								: "Save"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			{/* Delete Confirmation Modal */}
			<Dialog open={isDeleteModalOpen} onOpenChange={onClose}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Confirm Delete</DialogTitle>
						<DialogDescription>
							Are you sure you want to delete all sets for "
							{savedSet.MediaItem.Title}"? This action cannot be
							undone.
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button variant="outline" onClick={onClose}>
							Cancel
						</Button>
						<Button variant="destructive" onClick={confirmDelete}>
							Delete
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</Card>
	);
};

export default SavedSetsCard;
