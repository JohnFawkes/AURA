"use client";

import { Check } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";

type LogsFilterProps = {
	levelsFilter: string[];
	setLevelsFilter: (levels: string[]) => void;
	statusFilter: string[];
	setStatusFilter: (status: string[]) => void;
	actionsOptions: { label: string; value: string; section: string }[];
	actionsFilter: string[];
	setActionsFilter: (actions: string[]) => void;
};

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

export function LogsFilter({
	levelsFilter,
	setLevelsFilter,
	statusFilter,
	setStatusFilter,
	actionsOptions,
	actionsFilter,
	setActionsFilter,
}: LogsFilterProps) {
	return (
		<div className="flex-grow space-y-4 overflow-y-auto px-4 py-2">
			{/* Log Levels Filter */}
			<Label className="text-md font-semibold mb-2 block">Log Levels</Label>
			<div className="flex flex-col gap-1 max-h-56 overflow-y-auto border p-2 rounded-md">
				<div
					className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
						levelsFilter.length === 0 ? "bg-muted" : "hover:bg-muted/60"
					}`}
					onClick={() => {
						setLevelsFilter([]);
					}}
				>
					<Checkbox checked={levelsFilter.length === 0} id={`levels-all`} />
					<Label
						htmlFor={`levels-all`}
						className="text-sm flex-1 cursor-pointer truncate"
						onClick={(e) => e.stopPropagation()}
					>
						All Levels
					</Label>
					{levelsFilter.length === 0 && <Check className="h-4 w-4 text-primary" />}
				</div>
				{levelsOptions.map((level) => (
					<div
						key={level.value}
						className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
							levelsFilter.includes(level.value) ? "bg-muted" : "hover:bg-muted/60"
						}`}
						onClick={() => {
							let newLevels;
							if (levelsFilter.includes(level.value)) {
								newLevels = levelsFilter.filter((l) => l !== level.value);
							} else {
								newLevels = [...levelsFilter, level.value];
							}
							setLevelsFilter(newLevels);
						}}
					>
						<Checkbox checked={levelsFilter.includes(level.value)} id={`level-${level.value}`} />
						<Label
							htmlFor={`level-${level.value}`}
							className="text-sm flex-1 cursor-pointer truncate"
							onClick={(e) => e.stopPropagation()}
						>
							{level.label}
						</Label>
						{levelsFilter.includes(level.value) && <Check className="h-4 w-4 text-primary" />}
					</div>
				))}
			</div>
			<Separator className="my-4 w-full" />
			{/* Status Filter */}
			<Label className="text-md font-semibold mb-2 block">Status</Label>
			<div className="flex flex-col gap-1 max-h-56 overflow-y-auto border p-2 rounded-md">
				<div
					className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
						statusFilter.length === 0 ? "bg-muted" : "hover:bg-muted/60"
					}`}
					onClick={() => {
						setStatusFilter([]);
					}}
				>
					<Checkbox checked={statusFilter.length === 0} id={`status-all`} />
					<Label
						htmlFor={`status-all`}
						className="text-sm flex-1 cursor-pointer truncate"
						onClick={(e) => e.stopPropagation()}
					>
						All Statuses
					</Label>
					{statusFilter.length === 0 && <Check className="h-4 w-4 text-primary" />}
				</div>
				{statusOptions.map((status) => (
					<div
						key={status.value}
						className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
							statusFilter.includes(status.value) ? "bg-muted" : "hover:bg-muted/60"
						}`}
						onClick={() => {
							let newStatus;
							if (statusFilter.includes(status.value)) {
								newStatus = statusFilter.filter((s) => s !== status.value);
							} else {
								newStatus = [...statusFilter, status.value];
							}
							setStatusFilter(newStatus);
						}}
					>
						<Checkbox checked={statusFilter.includes(status.value)} id={`status-${status.value}`} />
						<Label
							htmlFor={`status-${status.value}`}
							className="text-sm flex-1 cursor-pointer truncate"
							onClick={(e) => e.stopPropagation()}
						>
							{status.label}
						</Label>
						{statusFilter.includes(status.value) && <Check className="h-4 w-4 text-primary" />}
					</div>
				))}
			</div>
			<Separator className="my-4 w-full" />

			{/* Actions Filter */}
			<Label className="text-md font-semibold mb-2 block">Actions</Label>
			<div className="flex flex-col gap-1 max-h-56 overflow-y-auto border p-2 rounded-md">
				<div
					className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
						actionsFilter.length === 0 ? "bg-muted" : "hover:bg-muted/60"
					}`}
					onClick={() => {
						setActionsFilter([]);
					}}
				>
					<Checkbox checked={actionsFilter.length === 0} id={`actions-all`} />
					<Label
						htmlFor={`actions-all`}
						className="text-sm flex-1 cursor-pointer truncate"
						onClick={(e) => e.stopPropagation()}
					>
						All Actions
					</Label>
					{actionsFilter.length === 0 && <Check className="h-4 w-4 text-primary" />}
				</div>
				{/* Group actions by section */}
				{Array.from(
					new Set(actionsOptions.sort((a, b) => a.section.localeCompare(b.section)).map((a) => a.section))
				).map((section) => (
					<div key={section}>
						<Separator className="my-2 w-full" />
						<Label className="text-xs font-semibold mb-1 block uppercase text-muted-foreground">
							{section.replace(/_/g, " ")}
						</Label>
						{actionsOptions
							.filter((a) => a.section === section)
							.map((action) => (
								<div
									key={action.value}
									className={`flex items-center space-x-2 px-2 py-1 rounded cursor-pointer transition-colors ${
										actionsFilter.includes(action.value) ? "bg-muted" : "hover:bg-muted/60"
									}`}
									onClick={() => {
										let newActions;
										if (actionsFilter.includes(action.value)) {
											newActions = actionsFilter.filter((a) => a !== action.value);
										} else {
											newActions = [...actionsFilter, action.value];
										}
										setActionsFilter(newActions);
									}}
								>
									<Checkbox
										checked={actionsFilter.includes(action.value)}
										id={`action-${action.value}`}
									/>
									<Label
										htmlFor={`action-${action.value}`}
										className="text-sm flex-1 cursor-pointer truncate"
										onClick={(e) => e.stopPropagation()}
									>
										{action.label}
									</Label>
									{actionsFilter.includes(action.value) && <Check className="h-4 w-4 text-primary" />}
								</div>
							))}
					</div>
				))}
			</div>
		</div>
	);
}
