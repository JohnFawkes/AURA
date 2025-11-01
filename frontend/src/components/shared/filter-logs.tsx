"use client";

import { Check, Filter, View } from "lucide-react";

import { useEffect, useMemo, useState } from "react";

import { SelectItemsPerPage } from "@/components/shared/select-items-per-page";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
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

import { cn } from "@/lib/cn";

import { TYPE_ITEMS_PER_PAGE_OPTIONS } from "@/types/ui-options";

const levelsOptions = [
	{ label: "TRACE", value: "trace" },
	{ label: "DEBUG", value: "debug" },
	{ label: "INFO", value: "info" },
	{ label: "WARNING", value: "warn" },
	{ label: "ERROR", value: "error" },
];

const statusOptions = [
	{ label: "Success", value: "success" },
	{ label: "Warning", value: "warn" },
	{ label: "Error", value: "error" },
];

type LogsFilterProps = {
	levelsFilter: string[];
	setLevelsFilter: (levels: string[]) => void;

	statusFilter: string[];
	setStatusFilter: (status: string[]) => void;

	actionsOptions: Record<string, { label: string; section: string }>;
	actionsFilter: string[];
	setActionsFilter: (actions: string[]) => void;

	setCurrentPage: (page: number) => void;
	itemsPerPage: TYPE_ITEMS_PER_PAGE_OPTIONS;
	setItemsPerPage: (num: TYPE_ITEMS_PER_PAGE_OPTIONS) => void;
};

function LogsFilterContent({
	levelsFilter,
	setLevelsFilter,
	statusFilter,
	setStatusFilter,
	actionsOptions,
	actionsFilter,
	setActionsFilter,
	setCurrentPage,
	itemsPerPage,
	setItemsPerPage,
}: LogsFilterProps) {
	const [pendingLevelsFilter, setPendingLevelsFilter] = useState<string[]>(levelsFilter);
	const [pendingStatusFilter, setPendingStatusFilter] = useState<string[]>(statusFilter);
	const [pendingActionsFilter, setPendingActionsFilter] = useState<string[]>(actionsFilter);

	const handleResetFilters = () => {
		setPendingLevelsFilter([]);
		setPendingStatusFilter([]);
		setPendingActionsFilter([]);
		setLevelsFilter([]);
		setStatusFilter([]);
		setActionsFilter([]);
		setCurrentPage(1);
	};

	const handleApplyFilters = () => {
		setCurrentPage(1);
		setLevelsFilter(pendingLevelsFilter);
		setStatusFilter(pendingStatusFilter);
		setActionsFilter(pendingActionsFilter);
	};

	// Group actions by section for rendering
	const actionsGroupedBySection = useMemo(() => {
		const grouped: Record<string, { label: string; value: string }[]> = {};
		Object.entries(actionsOptions).forEach(([path, info]) => {
			const section = info.section;
			if (!grouped[section]) grouped[section] = [];
			grouped[section].push({ label: info.label, value: path });
		});
		return grouped;
	}, [actionsOptions]);

	return (
		<div className="flex-grow space-y-4 overflow-y-auto px-4 py-2">
			{/* Items Per Page Selection */}
			<div className="flex items-center mb-4">
				<SelectItemsPerPage
					setCurrentPage={setCurrentPage}
					itemsPerPage={itemsPerPage}
					setItemsPerPage={setItemsPerPage}
				/>
			</div>
			{/* Log Levels Filter */}
			<Label className="text-md font-semibold mb-2 block">Log Levels</Label>
			<div className="flex flex-col gap-1 max-h-56 overflow-y-auto border p-2 rounded-md">
				<div
					className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
						pendingLevelsFilter.length === 0 ? "bg-muted" : "hover:bg-muted/60"
					}`}
					onClick={() => {
						setPendingLevelsFilter([]);
					}}
				>
					<Checkbox checked={pendingLevelsFilter.length === 0} id={`levels-all`} />
					<Label
						htmlFor={`levels-all`}
						className="text-sm flex-1 cursor-pointer truncate"
						onClick={(e) => e.stopPropagation()}
					>
						All Levels
					</Label>
					{pendingLevelsFilter.length === 0 && <Check className="h-4 w-4 text-primary" />}
				</div>
				{levelsOptions.map((level) => (
					<div
						key={level.value}
						className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
							pendingLevelsFilter.includes(level.value) ? "bg-muted" : "hover:bg-muted/60"
						}`}
						onClick={() => {
							let newLevels;
							if (pendingLevelsFilter.includes(level.value)) {
								newLevels = pendingLevelsFilter.filter((l) => l !== level.value);
							} else {
								newLevels = [...pendingLevelsFilter, level.value];
							}
							setPendingLevelsFilter(newLevels);
						}}
					>
						<Checkbox checked={pendingLevelsFilter.includes(level.value)} id={`level-${level.value}`} />
						<Label
							htmlFor={`level-${level.value}`}
							className="text-sm flex-1 cursor-pointer truncate"
							onClick={(e) => e.stopPropagation()}
						>
							{level.label}
						</Label>
						{pendingLevelsFilter.includes(level.value) && <Check className="h-4 w-4 text-primary" />}
					</div>
				))}
			</div>
			<Separator className="my-4 w-full" />
			{/* Status Filter */}
			<Label className="text-md font-semibold mb-2 block">Status</Label>
			<div className="flex flex-col gap-1 max-h-56 overflow-y-auto border p-2 rounded-md">
				<div
					className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
						pendingStatusFilter.length === 0 ? "bg-muted" : "hover:bg-muted/60"
					}`}
					onClick={() => {
						setPendingStatusFilter([]);
					}}
				>
					<Checkbox checked={pendingStatusFilter.length === 0} id={`status-all`} />
					<Label
						htmlFor={`status-all`}
						className="text-sm flex-1 cursor-pointer truncate"
						onClick={(e) => e.stopPropagation()}
					>
						All Statuses
					</Label>
					{pendingStatusFilter.length === 0 && <Check className="h-4 w-4 text-primary" />}
				</div>
				{statusOptions.map((status) => (
					<div
						key={status.value}
						className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
							pendingStatusFilter.includes(status.value) ? "bg-muted" : "hover:bg-muted/60"
						}`}
						onClick={() => {
							let newStatus;
							if (pendingStatusFilter.includes(status.value)) {
								newStatus = pendingStatusFilter.filter((s) => s !== status.value);
							} else {
								newStatus = [...pendingStatusFilter, status.value];
							}
							setPendingStatusFilter(newStatus);
						}}
					>
						<Checkbox checked={pendingStatusFilter.includes(status.value)} id={`status-${status.value}`} />
						<Label
							htmlFor={`status-${status.value}`}
							className="text-sm flex-1 cursor-pointer truncate"
							onClick={(e) => e.stopPropagation()}
						>
							{status.label}
						</Label>
						{pendingStatusFilter.includes(status.value) && <Check className="h-4 w-4 text-primary" />}
					</div>
				))}
			</div>
			<Separator className="my-4 w-full" />

			{/* Actions Filter */}
			<Label className="text-md font-semibold mb-2 block">Actions</Label>
			<div className="flex flex-col gap-1 max-h-56 overflow-y-auto border p-2 rounded-md">
				<div
					className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
						pendingActionsFilter.length === 0 ? "bg-muted" : "hover:bg-muted/60"
					}`}
					onClick={() => {
						setPendingActionsFilter([]);
					}}
				>
					<Checkbox checked={pendingActionsFilter.length === 0} id={`actions-all`} />
					<Label
						htmlFor={`actions-all`}
						className="text-sm flex-1 cursor-pointer truncate"
						onClick={(e) => e.stopPropagation()}
					>
						All Actions
					</Label>
					{pendingActionsFilter.length === 0 && <Check className="h-4 w-4 text-primary" />}
				</div>
				{/* Group actions by section */}
				{Object.entries(actionsGroupedBySection)
					.sort(([a], [b]) => a.localeCompare(b))
					.map(([section, actions]) => (
						<div key={section}>
							<Separator className="my-2 w-full" />
							<Label className="text-xs font-semibold mb-1 block uppercase text-muted-foreground">
								{section.replace(/_/g, " ")}
							</Label>
							{actions.map((action) => (
								<div
									key={action.value}
									className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
										pendingActionsFilter.includes(action.value) ? "bg-muted" : "hover:bg-muted/60"
									}`}
									onClick={() => {
										let newActions;
										if (pendingActionsFilter.includes(action.value)) {
											newActions = pendingActionsFilter.filter((a) => a !== action.value);
										} else {
											newActions = [...pendingActionsFilter, action.value];
										}
										setPendingActionsFilter(newActions);
									}}
								>
									<Checkbox
										checked={pendingActionsFilter.includes(action.value)}
										id={`action-${action.value}`}
									/>
									<Label
										htmlFor={`action-${action.value}`}
										className="text-sm flex-1 cursor-pointer truncate"
										onClick={(e) => e.stopPropagation()}
									>
										{action.label}
									</Label>
									{pendingActionsFilter.includes(action.value) && (
										<Check className="h-4 w-4 text-primary" />
									)}
								</div>
							))}
						</div>
					))}
			</div>

			{/* Apply Filters Button */}
			<Button
				className="w-full mt-2 cursor-pointer hover:brightness-120 active:scale-95"
				onClick={handleApplyFilters}
			>
				Apply Filters
			</Button>

			{/* Reset Filters Button */}
			<Button
				variant={"destructive"}
				className="w-full cursor-pointer hover:brightness-120 active:scale-95"
				onClick={handleResetFilters}
			>
				Reset Filters
			</Button>
		</div>
	);
}

export function FilterLogs({
	levelsFilter,
	setLevelsFilter,
	statusFilter,
	setStatusFilter,
	actionsOptions,
	actionsFilter,
	setActionsFilter,
	setCurrentPage,
	itemsPerPage,
	setItemsPerPage,
}: LogsFilterProps) {
	const [isWideScreen, setIsWideScreen] = useState(typeof window !== "undefined" ? window.innerWidth >= 1300 : false);

	// Change isWideScreen on window resize
	useEffect(() => {
		const handleResize = () => {
			setIsWideScreen(window.innerWidth >= 1300);
		};
		handleResize();
		window.addEventListener("resize", handleResize);
		return () => window.removeEventListener("resize", handleResize);
	}, []);

	// Calculate number of active filters
	const numberOfActiveFilters = useMemo(() => {
		let count = 0;
		if (levelsFilter.length > 0) count++;
		if (statusFilter.length > 0) count++;
		if (actionsFilter.length > 0) count++;
		return count;
	}, [levelsFilter, statusFilter, actionsFilter]);

	return (
		<>
			{isWideScreen ? (
				<Popover>
					<PopoverTrigger asChild>
						<div>
							<Button
								variant="outline"
								className={cn(numberOfActiveFilters > 0 && "ring-1 ring-primary ring-offset-1")}
							>
								<View className="h-5 w-5" />
								View & Filter {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
								<Filter className="h-5 w-5" />
							</Button>
						</div>
					</PopoverTrigger>
					<PopoverContent
						side="right"
						align="start"
						className="w-[350px] p-2 bg-background border border-primary"
					>
						<LogsFilterContent
							levelsFilter={levelsFilter}
							setLevelsFilter={setLevelsFilter}
							statusFilter={statusFilter}
							setStatusFilter={setStatusFilter}
							actionsOptions={actionsOptions}
							actionsFilter={actionsFilter}
							setActionsFilter={setActionsFilter}
							setCurrentPage={setCurrentPage}
							itemsPerPage={itemsPerPage}
							setItemsPerPage={setItemsPerPage}
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
							<View className="h-5 w-5" />
							View & Filter {numberOfActiveFilters > 0 && `(${numberOfActiveFilters})`}
							<Filter className="h-5 w-5" />
						</Button>
					</DrawerTrigger>
					<DrawerContent>
						<DrawerHeader className="my-0">
							<DrawerTitle className="mb-0">View & Filter</DrawerTitle>
							<DrawerDescription className="mb-0">
								Use the options below to change the number of logs displayed and filter logs by level,
								status, or action/path.
							</DrawerDescription>
						</DrawerHeader>
						<Separator className="my-1 w-full" />
						<LogsFilterContent
							levelsFilter={levelsFilter}
							setLevelsFilter={setLevelsFilter}
							statusFilter={statusFilter}
							setStatusFilter={setStatusFilter}
							actionsOptions={actionsOptions}
							actionsFilter={actionsFilter}
							setActionsFilter={setActionsFilter}
							setCurrentPage={setCurrentPage}
							itemsPerPage={itemsPerPage}
							setItemsPerPage={setItemsPerPage}
						/>
					</DrawerContent>
				</Drawer>
			)}
		</>
	);
}
