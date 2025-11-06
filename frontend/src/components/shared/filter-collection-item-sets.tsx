"use client";

import { Filter } from "lucide-react";

import { useState } from "react";

import { PopoverHelp } from "@/components/shared/popover-help";
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

import { cn } from "@/lib/cn";

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
	return (
		<div className="flex-grow space-y-4 overflow-y-auto px-4 py-2">
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
