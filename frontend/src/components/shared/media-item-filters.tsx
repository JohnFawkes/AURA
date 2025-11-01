"use client";

import { Filter } from "lucide-react";

import { useEffect, useState } from "react";

import Link from "next/link";

import { PopoverHelp } from "@/components/shared/popover-help";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Drawer,
	DrawerContent,
	DrawerDescription,
	DrawerHeader,
	DrawerTitle,
	DrawerTrigger,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { ToggleGroup } from "@/components/ui/toggle-group";

import { cn } from "@/lib/cn";
import { useUserPreferencesStore } from "@/lib/stores/global-user-preferences";

import { DOWNLOAD_DEFAULT_TYPE_OPTIONS, TYPE_DOWNLOAD_DEFAULT_OPTIONS } from "@/types/ui-options";

type MediaItemFilterProps = {
	numberOfActiveFilters: number;
	hiddenCount: number;
	showHiddenUsers: boolean;
	handleShowHiddenUsers: (val: boolean) => void;
	hasTitleCards: boolean;
	showOnlyTitlecardSets: boolean;
	handleShowSetsWithTitleCardsOnly: (val: boolean) => void;
};

export function MediaItemFilter({
	numberOfActiveFilters,
	hiddenCount,
	showHiddenUsers,
	handleShowHiddenUsers,
	hasTitleCards,
	showOnlyTitlecardSets,
	handleShowSetsWithTitleCardsOnly,
}: MediaItemFilterProps) {
	// Is Wide Screen State
	const [isWideScreen, setIsWideScreen] = useState(false);

	useEffect(() => {
		const handleResize = () => {
			if (window.innerWidth < 1300) {
				setIsWideScreen(false);
			} else {
				setIsWideScreen(true);
			}
		};
		handleResize();
		window.addEventListener("resize", handleResize);
		return () => window.removeEventListener("resize", handleResize);
	}, []);

	return (
		<div>
			{isWideScreen ? (
				<Popover>
					<PopoverTrigger asChild>
						<div>
							<Button
								variant="outline"
								className={cn(numberOfActiveFilters > 0 && "ring-2 ring-primary")}
							>
								Filters {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
								<Filter className="h-5 w-5" />
							</Button>
						</div>
					</PopoverTrigger>
					<PopoverContent
						side="right"
						align="start"
						className="w-[350px] p-2 bg-background border border-primary"
					>
						<MediaItemFilterContent
							numberOfActiveFilters={numberOfActiveFilters}
							hiddenCount={hiddenCount}
							showHiddenUsers={showHiddenUsers}
							handleShowHiddenUsers={handleShowHiddenUsers}
							hasTitleCards={hasTitleCards}
							showOnlyTitlecardSets={showOnlyTitlecardSets}
							handleShowSetsWithTitleCardsOnly={handleShowSetsWithTitleCardsOnly}
						/>
					</PopoverContent>
				</Popover>
			) : (
				<Drawer direction="left">
					<DrawerTrigger asChild>
						<Button
							variant="outline"
							className={cn(numberOfActiveFilters > 0 && "ring-1 ring-primary ring-offset-1")}
						>
							Filters {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
							<Filter className="h-5 w-5" />
						</Button>
					</DrawerTrigger>
					<DrawerContent>
						<DrawerHeader className="my-0">
							<DrawerTitle className="mb-0">Filters</DrawerTitle>
							<DrawerDescription className="mb-0">
								Use the options below to filter the poster sets.
							</DrawerDescription>
						</DrawerHeader>
						<Separator className="my-1 w-full" />
						<MediaItemFilterContent
							numberOfActiveFilters={numberOfActiveFilters}
							hiddenCount={hiddenCount}
							showHiddenUsers={showHiddenUsers}
							handleShowHiddenUsers={handleShowHiddenUsers}
							hasTitleCards={hasTitleCards}
							showOnlyTitlecardSets={showOnlyTitlecardSets}
							handleShowSetsWithTitleCardsOnly={handleShowSetsWithTitleCardsOnly}
						/>
					</DrawerContent>
				</Drawer>
			)}
		</div>
	);
}

function MediaItemFilterContent({
	hiddenCount,
	showHiddenUsers,
	handleShowHiddenUsers,
	hasTitleCards,
	showOnlyTitlecardSets,
	handleShowSetsWithTitleCardsOnly,
}: MediaItemFilterProps) {
	const downloadDefaultsTypes = useUserPreferencesStore((state) => state.downloadDefaults);
	const setDownloadDefaultsTypes = useUserPreferencesStore((state) => state.setDownloadDefaults);
	const showonlyDownloadDefaults = useUserPreferencesStore((state) => state.showOnlyDownloadDefaults);
	const setShowOnlyDownloadDefaults = useUserPreferencesStore((state) => state.setShowOnlyDownloadDefaults);

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
				<div className="flex items-center space-x-2 justify-between mt-4">
					<div className="flex items-center space-x-2">
						<Switch
							className="ml-0"
							checked={showonlyDownloadDefaults}
							onCheckedChange={() => setShowOnlyDownloadDefaults(!showonlyDownloadDefaults)}
						/>{" "}
						<Label>Only show selected image types</Label>
					</div>

					<PopoverHelp ariaLabel="help-filter-image-types">
						<p className="mb-2">
							If checked, only sets that contain at least one of the selected image types will be shown.
						</p>
						<p className="text-muted-foreground">
							This is global setting that will be applied to all media items and user sets. You can always
							change this setting in this filter, or in the Settings Page{" "}
							<Link href="/settings#preferences-section" className="underline">
								User Preferences
							</Link>{" "}
							Section.
						</p>
					</PopoverHelp>
				</div>

				{/* Hidden Users*/}
				{hiddenCount > 0 && (
					<>
						<Separator className="my-4 w-full" />
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

				{/* Mandatory Titlecard Sets */}
				{hasTitleCards && (!showonlyDownloadDefaults || downloadDefaultsTypes.includes("titlecard")) && (
					<>
						<Separator className="my-4 w-full" />
						<Label className="text-md font-semibold mb-1 block">Titlecard Filter</Label>
						<div className="justify-between flex items-center">
							<div className="flex items-center space-x-2">
								<Switch
									className="ml-0"
									checked={showOnlyTitlecardSets}
									onCheckedChange={handleShowSetsWithTitleCardsOnly}
								/>
								<Label>Only show sets with titlecards</Label>
							</div>
							<PopoverHelp ariaLabel="media-item-filter-titlecards">
								<p className="mb-2">
									When enabled, only sets that contain titlecards will be shown in the list.
								</p>
							</PopoverHelp>
						</div>
					</>
				)}
				<Separator className="my-4 w-full" />
			</div>
		</div>
	);
}
