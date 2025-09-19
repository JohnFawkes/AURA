"use client";

import { Plus } from "lucide-react";

import React, { useEffect, useRef } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import { AppConfigKometa } from "@/types/config/config-app";

interface ConfigSectionKometaProps {
	value: AppConfigKometa;
	editing: boolean;
	dirtyFields?: Partial<Record<keyof AppConfigKometa, boolean>>;
	onChange: <K extends keyof AppConfigKometa>(field: K, value: AppConfigKometa[K]) => void;
	errorsUpdate?: (errors: Partial<Record<keyof AppConfigKometa, string>>) => void;
}

export const ConfigSectionKometa: React.FC<ConfigSectionKometaProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
}) => {
	const prevErrorsRef = useRef<string>("");

	// Safely normalize Labels (avoid null/undefined errors)
	const labels = React.useMemo(
		() => (Array.isArray(value.Labels) ? value.Labels.filter((l): l is string => typeof l === "string") : []),
		[value.Labels]
	);

	// Validation: only when Kometa is enabled
	const errors = React.useMemo(() => {
		const errs: Partial<Record<keyof AppConfigKometa, string>> = {};

		if (value.RemoveLabels) {
			// Only enforce label rules when enabled
			if (labels.length === 0) {
				errs.Labels = "Labels cannot be empty.";
			} else if (labels.some((l) => !l.trim())) {
				errs.Labels = "Labels cannot be empty.";
			} else {
				const seen = new Set<string>();
				for (const l of labels) {
					const lower = l.trim().toLowerCase();
					if (seen.has(lower)) {
						errs.Labels = "Duplicate labels not allowed.";
						break;
					}
					seen.add(lower);
				}
			}
		}

		return errs;
	}, [value.RemoveLabels, labels]);

	useEffect(() => {
		if (!errorsUpdate) return;
		const ser = JSON.stringify(errors);
		if (ser === prevErrorsRef.current) return;
		prevErrorsRef.current = ser;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	// Label helpers
	const addLabel = (raw: string) => {
		const name = raw.trim();
		if (!name) return;
		if (labels.some((l) => l.trim().toLowerCase() === name.toLowerCase())) return;
		onChange("Labels", [...labels, name] as string[]);
	};
	const removeLabelAt = (idx: number) => {
		const next = labels.slice();
		next.splice(idx, 1);
		onChange("Labels", next as string[]);
	};
	const newLabelRef = useRef<HTMLInputElement | null>(null);

	return (
		<Card className="p-5 space-y-1">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Kometa</h2>
			</div>

			{/* RemoveLabels toggle */}
			<div
				className={cn(
					"flex items-center justify-between border rounded-md p-3 transition",
					"border-muted",
					dirtyFields.RemoveLabels && "border-amber-500"
				)}
			>
				<Label>Remove Labels</Label>
				<div className="flex items-center gap-2">
					<Switch
						disabled={!editing}
						checked={value.RemoveLabels}
						onCheckedChange={(c) => onChange("RemoveLabels", c)}
					/>
					{editing && (
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-kometa-remove-labels"
								>
									?
								</Button>
							</PopoverTrigger>
							<PopoverContent side="top" align="end" sideOffset={6} className="w-64 text-xs leading-snug">
								<p className="mb-2">
									When enabled, existing labels managed by Kometa are removed before applying the
									current Labels list.
								</p>
								<p className="text-[10px] text-muted-foreground">
									Disable to only add / ensure labels without removing others.
								</p>
							</PopoverContent>
						</Popover>
					)}
				</div>
			</div>

			{/* Labels list */}
			<div
				className={cn(
					"space-y-3",
					(dirtyFields.Labels || errors.Labels) && "rounded-md",
					errors.Labels ? "border border-red-500 p-3" : dirtyFields.Labels && "border border-amber-500 p-3"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>Labels</Label>
					{editing && (
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-kometa-labels"
								>
									?
								</Button>
							</PopoverTrigger>
							<PopoverContent
								side="right"
								align="center"
								sideOffset={8}
								className="w-72 text-xs leading-snug"
							>
								<p className="mb-2">
									List of labels to remove from Plex. Click a badge to remove it while editing.
								</p>
								<p className="text-[10px] text-muted-foreground">
									Undefined or null list is treated as empty.
								</p>
							</PopoverContent>
						</Popover>
					)}
				</div>

				<div className="flex flex-wrap items-center gap-2">
					{labels.length === 0 && (
						<span className="inline-flex h-7 items-center rounded-full border border-dashed px-3 text-[11px] text-muted-foreground">
							No labels added
						</span>
					)}
					{labels.map((label, i) => (
						<Badge
							key={i}
							className={cn(
								"cursor-pointer select-none transition duration-200 px-3 py-1 text-[11px] font-normal",
								editing
									? "bg-secondary text-secondary-foreground hover:bg-red-500 hover:text-white"
									: "bg-muted text-muted-foreground"
							)}
							onClick={() => {
								if (!editing) return;
								removeLabelAt(i);
							}}
							title={editing ? "Remove label" : label}
						>
							{label}
						</Badge>
					))}
					{editing && (
						<form
							onSubmit={(e) => {
								e.preventDefault();
								if (!newLabelRef.current) return;
								addLabel(newLabelRef.current.value);
								newLabelRef.current.value = "";
							}}
							className="flex items-center gap-1"
						>
							<Input
								ref={newLabelRef}
								placeholder="Add label..."
								className="h-7 w-40 text-xs"
								onKeyDown={(e) => {
									if (e.key === "Enter") {
										e.preventDefault();
										const target = e.currentTarget;
										addLabel(target.value);
										target.value = "";
									}
								}}
							/>
							<Button type="submit" variant="outline" size="icon" className="h-7 w-7" title="Add label">
								<Plus className="h-4 w-4" />
							</Button>
						</form>
					)}
				</div>

				{errors.Labels && <p className="text-xs text-red-500">{errors.Labels}</p>}
			</div>
		</Card>
	);
};
