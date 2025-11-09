"use client";

import { Filter } from "lucide-react";

import { useState } from "react";

import { PopoverHelp } from "@/components/shared/popover-help";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { ToggleGroup } from "@/components/ui/toggle-group";

import { cn } from "@/lib/cn";
import { useUserPreferencesStore } from "@/lib/stores/global-user-preferences";

import { DOWNLOAD_DEFAULT_TYPE_OPTIONS, TYPE_DOWNLOAD_DEFAULT_OPTIONS } from "@/types/ui-options";

type CollectionItemFilterProps = {
	numberOfActiveFilters?: number;
	hiddenCount: number;
	showHiddenUsers: boolean;
	handleShowHiddenUsers: (val: boolean) => void;
};

function CollectionItemFilterContent({
	hiddenCount,
	showHiddenUsers,
	handleShowHiddenUsers,
}: CollectionItemFilterProps) {
	const downloadDefaultsTypes = useUserPreferencesStore((state) => state.downloadDefaults);
	const setDownloadDefaultsTypes = useUserPreferencesStore((state) => state.setDownloadDefaults);

	return (
		<div className="flex-grow space-y-4 overflow-y-auto px-4 py-2">
			<div className="flex flex-col">
				{/* Download Defaults */}
				<div className="flex items-center space-x-2 justify-between">
					<Label className="text-md font-semibold block">Download Defaults</Label>

					<PopoverHelp ariaLabel="help-default-image-types">
						<p className="mb-2">
							Select which image types you want auto-checked for each download. This will let you avoid
							unchecking them manually for each download.
						</p>
						<p className="text-muted-foreground">Click a badge to toggle it on or off.</p>
					</PopoverHelp>
				</div>
				<ToggleGroup
					type="multiple"
					className="flex flex-wrap gap-2 ml-2 mt-2"
					value={downloadDefaultsTypes}
					onValueChange={(value: TYPE_DOWNLOAD_DEFAULT_OPTIONS[]) => {
						// Ensure at least one type is always selected
						if (value.length === 0) return;
						setDownloadDefaultsTypes(value);
					}}
				>
					{DOWNLOAD_DEFAULT_TYPE_OPTIONS.map((type) => (
						<Badge
							key={type}
							className={cn(
								"cursor-pointer text-sm px-3 py-1 font-normal transition active:scale-95",
								downloadDefaultsTypes.includes(type)
									? "bg-primary text-primary-foreground hover:brightness-120"
									: "bg-muted text-muted-foreground border hover:text-accent-foreground"
							)}
							variant={downloadDefaultsTypes.includes(type) ? "default" : "outline"}
							onClick={() => {
								if (downloadDefaultsTypes.includes(type)) {
									// Only allow removal if more than one type is selected
									if (downloadDefaultsTypes.length > 1) {
										setDownloadDefaultsTypes(downloadDefaultsTypes.filter((t) => t !== type));
									}
								} else {
									setDownloadDefaultsTypes([...downloadDefaultsTypes, type]);
								}
							}}
							style={
								downloadDefaultsTypes.includes(type) && downloadDefaultsTypes.length === 1
									? { opacity: 0.5, pointerEvents: "none" }
									: undefined
							}
						>
							{type.charAt(0).toUpperCase() + type.slice(1).replace(/([A-Z])/g, " $1")}
						</Badge>
					))}
				</ToggleGroup>
			</div>
			<div className="flex flex-col">
				{/* Hidden Users*/}
				{hiddenCount > 0 && (
					<>
						<Label className="text-md font-semibold mb-1 block">Hidden Users</Label>
						<div className="justify-between flex items-center">
							<div className="flex items-center space-x-2">
								<Switch
									className="ml-0"
									checked={showHiddenUsers}
									onCheckedChange={handleShowHiddenUsers}
									disabled={hiddenCount === 0}
								/>{" "}
								<Label>Show hidden users</Label>
							</div>
							<PopoverHelp ariaLabel="media-item-filter-hidden-users">
								<p className="mb-2">
									When enabled, sets from users you have hidden will be shown in the list.
								</p>
								<p className="text-muted-foreground">You can hide users directly in the MediUX site.</p>
							</PopoverHelp>
						</div>
					</>
				)}
				<Separator className="my-4 w-full" />
			</div>
		</div>
	);
}

export function CollectionItemFilter({
	numberOfActiveFilters = 0,
	hiddenCount,
	showHiddenUsers,
	handleShowHiddenUsers,
}: CollectionItemFilterProps) {
	// State - Open/Close Modal
	const [modalOpen, setModalOpen] = useState(false);

	return (
		<Dialog open={modalOpen} onOpenChange={setModalOpen}>
			<DialogTrigger asChild>
				<Button
					variant="outline"
					className={cn(numberOfActiveFilters > 0 && "ring-1 ring-primary ring-offset-1")}
				>
					Filters {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
					<Filter className="h-5 w-5" />
				</Button>
			</DialogTrigger>
			<DialogContent
				className={cn("z-50", "max-h-[80vh] overflow-y-auto", "sm:max-w-[700px]", "border border-primary")}
			>
				<DialogHeader>
					<DialogTitle>Filters</DialogTitle>
					<DialogDescription>Use the options below to filter the collection item sets.</DialogDescription>
				</DialogHeader>
				<Separator className="my-1 w-full" />
				<CollectionItemFilterContent
					hiddenCount={hiddenCount}
					showHiddenUsers={showHiddenUsers}
					handleShowHiddenUsers={handleShowHiddenUsers}
				/>
			</DialogContent>
		</Dialog>
	);
}
