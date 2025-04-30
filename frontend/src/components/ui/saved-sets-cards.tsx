"use client";
import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardHeader,
	CardContent,
	CardFooter,
} from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogFooter,
	DialogTitle,
} from "@/components/ui/dialog";
import { H4, P, Small } from "@/components/ui/typography";
import { deleteItemFromDB, patchSelectedTypesInDB } from "@/services/api.db";
import { ClientMessage } from "@/types/clientMessage";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "./dropdown-menu";
import { Badge } from "./badge";
import { Separator } from "./separator";
import { CheckCircle2 as Checkmark } from "lucide-react";
import { X } from "lucide-react";
import Image from "next/image";
import { MoreHorizontal } from "lucide-react";
import { log } from "@/lib/logger";
import { DialogDescription } from "@radix-ui/react-dialog";
import Link from "next/link";
import { cn } from "@/lib/utils";

const formatDate = (dateString: string) => {
	try {
		const date = new Date(dateString);
		return new Intl.DateTimeFormat("en-US", {
			year: "numeric",
			month: "long",
			day: "numeric",
			hour: "2-digit",
			minute: "2-digit",
		}).format(date);
	} catch {
		return "Invalid Date";
	}
};

const SavedSetsCard: React.FC<{
	id: string;
	savedSet: ClientMessage;
	onUpdate: () => void;
}> = ({ id, savedSet, onUpdate }) => {
	const [isEditModalOpen, setIsEditModalOpen] = useState(false);
	const [editSelectedTypes, setEditSelectedTypes] = useState<string[]>(
		savedSet.SelectedTypes
	);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [updateError, setUpdateError] = useState("");

	const confirmEdit = async () => {
		const resp = await patchSelectedTypesInDB(id, editSelectedTypes);
		if (resp.status !== "success") {
			setUpdateError(resp.message);
		} else {
			setUpdateError("");
			setIsEditModalOpen(false);
			onUpdate();
		}
	};

	const confirmDelete = async () => {
		const resp = await deleteItemFromDB(id);
		if (resp.status !== "success") {
			setUpdateError(resp.message);
		} else {
			setIsDeleteModalOpen(false);
			setUpdateError("");
			onUpdate();
		}
	};

	const renderBadges = () =>
		savedSet.SelectedTypes.map((type) => (
			<Badge key={type}>
				{type === "poster"
					? "Poster"
					: type === "backdrop"
					? "Backdrop"
					: type === "seasonPoster"
					? "Season Posters"
					: type === "titlecard"
					? "Title Card"
					: type}
			</Badge>
		));

	log("SavedSetsCard - Rendered", savedSet);

	return (
		<Card className="relative w-full max-w-md mx-auto mb-4">
			<CardHeader>
				{/* Top Left: Auto Download Icon */}
				<div className="absolute top-2 left-2">
					{savedSet.AutoDownload ? (
						<Checkmark className="text-green-500" size={24} />
					) : (
						<X className="text-red-500" size={24} />
					)}
				</div>

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
				<P
					className={cn(
						"text-sm text-muted-foreground",
						savedSet.MediaItem.Year
					)}
				>
					Year: {savedSet.MediaItem.Year}
				</P>

				{/* Library Title */}
				<P
					className={cn(
						"text-sm text-muted-foreground",
						savedSet.MediaItem.Year
					)}
				>
					Library: {savedSet.MediaItem.LibraryTitle}
				</P>

				{/* Last Updated */}
				<P
					className={cn(
						"text-sm text-muted-foreground",
						savedSet.MediaItem.Year
					)}
				>
					Last Updated: {formatDate(savedSet.LastUpdate || "")}
				</P>

				<Link
					href={`https://mediux.pro/sets/${savedSet.Set.ID}`}
					target="_blank"
					rel="noopener noreferrer"
					className="text-primary font-medium hover:underline"
				>
					View Set: {savedSet.Set.ID}
				</Link>

				{/* Separator */}
				<Separator className="my-4" />

				{/* Badges */}
				<div className="flex flex-wrap gap-2">{renderBadges()}</div>
			</CardContent>

			{/* Edit Modal */}
			<Dialog open={isEditModalOpen} onOpenChange={setIsEditModalOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Edit Saved Set</DialogTitle>
						<DialogDescription>
							Select the types you want to include in the saved
							set.
						</DialogDescription>
					</DialogHeader>
					<div className="flex flex-wrap gap-2">
						{[
							"poster",
							"backdrop",
							"seasonPoster",
							"titlecard",
						].map((type) => (
							<Badge
								key={type}
								className="flex items-center gap-2"
								variant={
									editSelectedTypes.includes(type)
										? "default"
										: "outline"
								}
								onClick={() => {
									setEditSelectedTypes((prev) =>
										prev.includes(type)
											? prev.filter((t) => t !== type)
											: [...prev, type]
									);
								}}
							>
								{type === "poster"
									? "Poster"
									: type === "backdrop"
									? "Backdrop"
									: type === "seasonPoster"
									? "Season Posters"
									: type === "titlecard"
									? "Title Card"
									: type}
							</Badge>
						))}
					</div>
					{updateError && (
						<Small className="text-destructive">
							{updateError}
						</Small>
					)}
					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => setIsEditModalOpen(false)}
						>
							Cancel
						</Button>
						<Button onClick={confirmEdit}>Save</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			{/* Delete Confirmation Modal */}
			<Dialog
				open={isDeleteModalOpen}
				onOpenChange={setIsDeleteModalOpen}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Confirm Delete</DialogTitle>
						<DialogDescription>
							Are you sure you want to delete this saved set? This
							action cannot be undone.
						</DialogDescription>
					</DialogHeader>

					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => setIsDeleteModalOpen(false)}
						>
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
