"use client";

import { HelpCircle, Plus, Trash2 } from "lucide-react";

import React, { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import { AppConfigLabelsAndTags, AppConfigLabelsAndTagsApplication } from "@/types/config/config-app";

interface ConfigSectionLabelsAndTagsProps {
	value: AppConfigLabelsAndTags;
	editing: boolean;
	dirtyFields?: {
		Applications?: Array<Partial<Record<string, boolean | { Enabled?: boolean; Add?: boolean; Remove?: boolean }>>>;
	};
	onChange: <K extends keyof AppConfigLabelsAndTags>(field: K, value: AppConfigLabelsAndTags[K]) => void;
	errorsUpdate?: (errors: Record<string, string>) => void;
}

const APPLICATION_TYPES = ["Plex"];

export const ConfigSectionLabelsAndTags: React.FC<ConfigSectionLabelsAndTagsProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
}) => {
	const prevErrorsRef = useRef<string>("");

	const [newApplicationType, setNewApplicationType] = useState("Plex");

	const applications = React.useMemo(
		() => (Array.isArray(value.Applications) ? value.Applications : []),
		[value.Applications]
	);

	// ----- Validation -----
	const errors = React.useMemo(() => {
		const errs: Record<string, string> = {};

		applications.forEach((app, idx) => {
			const addLabels = Array.isArray(app.Add) ? app.Add : [];
			const removeLabels = Array.isArray(app.Remove) ? app.Remove : [];
			if (app.Enabled && addLabels.length === 0 && removeLabels.length === 0) {
				errs[`Applications.${idx}`] = "At least one label to add or remove must be specified.";
			}
			const intersection = addLabels.filter((label) => removeLabels.includes(label));
			if (intersection.length > 0) {
				errs[`Applications.${idx}`] = `Labels cannot be in both Add and Remove: ${intersection.join(", ")}`;
			}
		});

		return errs;
	}, [applications]);

	useEffect(() => {
		if (!errorsUpdate) return;
		const serialized = JSON.stringify(errors);
		if (serialized === prevErrorsRef.current) return;
		prevErrorsRef.current = serialized;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	// ----- Mutators -----
	const setApplications = (apps: AppConfigLabelsAndTagsApplication[]) => onChange("Applications", apps);
	const providerExists = applications.some((app) => app.Application === newApplicationType);

	const addApplication = () => {
		if (!editing || providerExists) return;
		const type = newApplicationType;
		let newEntry: AppConfigLabelsAndTagsApplication;
		if (type === "Plex") {
			newEntry = { Application: "Plex", Enabled: true, Add: [], Remove: [] };
		}
		setApplications([...applications, newEntry!]);
	};

	const removeApplication = (index: number) => {
		if (!editing) return;
		const apps = [...applications];
		apps.splice(index, 1);
		setApplications(apps);
	};

	const updateEnabled = (index: number, enabled: boolean) => {
		const next = applications.slice();
		next[index] = { ...next[index], Enabled: enabled };
		setApplications(next);
	};

	const updateApplication = <K extends keyof AppConfigLabelsAndTagsApplication>(
		index: number,
		field: K,
		val: AppConfigLabelsAndTagsApplication[K]
	) => {
		const next = applications.slice();
		next[index] = { ...next[index], [field]: val };
		setApplications(next);
	};

	return (
		<Card className="p-5 space-y-6">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Labels & Tags</h2>
			</div>
			<div
				className={cn(
					"space-y-4",
					(errors["Applications"] || dirtyFields.Applications) && "rounded-md",
					errors["Applications"]
						? "border border-red-500 p-3"
						: dirtyFields.Applications && "border border-amber-500 p-3"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>Applications</Label>
					{editing && (
						<div className="flex items-center gap-2">
							<Select value={newApplicationType} onValueChange={(v) => setNewApplicationType(v)}>
								<SelectTrigger className="h-8 w-36">
									<SelectValue placeholder="Type" />
								</SelectTrigger>
								<SelectContent>
									{APPLICATION_TYPES.filter(
										(p) => !applications.some((app) => app.Application === p)
									).map((p) => (
										<SelectItem key={p} value={p}>
											{p}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
							<Button
								type="button"
								variant="outline"
								size="sm"
								onClick={addApplication}
								disabled={providerExists}
							>
								<Plus className="h-4 w-4 mr-1" />
								Add
							</Button>
						</div>
					)}
				</div>

				{applications.length === 0 && (
					<p className="text-[11px] text-muted-foreground">No applications configured.</p>
				)}

				{applications.map((app, idx) => {
					const appDirty = dirtyFields.Applications?.[idx] as Partial<{
						Enabled?: boolean;
						Add?: boolean;
						Remove?: boolean;
					}>;
					const appError = errors[`Applications.${idx}`];
					const addLabels = Array.isArray(app.Add) ? app.Add : [];
					const removeLabels = Array.isArray(app.Remove) ? app.Remove : [];

					return (
						<div
							key={idx}
							className={cn(
								"space-y-3 rounded-md border p-3 transition",
								appError ? "border-red-500" : appDirty ? "border-amber-500" : "border-muted"
							)}
						>
							<div className="flex items-center justify-between">
								<div className="flex items-center gap-3">
									<p className="font-medium text-sm">{app.Application}</p>
									<Switch
										disabled={!editing}
										checked={app.Enabled}
										onCheckedChange={(v) => updateEnabled(idx, v)}
									/>
								</div>
								{editing && (
									<Button
										variant="ghost"
										size="icon"
										onClick={() => removeApplication(idx)}
										aria-label="remove-application"
										disabled={applications.length === 1}
									>
										<Trash2 className="h-4 w-4" />
									</Button>
								)}
							</div>

							{/* Add Labels */}
							<div className="space-y-2">
								<div className="flex items-center justify-between">
									<Label>Add Labels</Label>
									{editing && (
										<Popover>
											<PopoverTrigger asChild>
												<Button
								variant="outline"
								className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
								aria-label="help-labels-to-add"
							>
								<HelpCircle className="h-4 w-4" />
							</Button>
											</PopoverTrigger>
											<PopoverContent
												side="right"
												align="center"
												sideOffset={8}
												className="w-72 text-xs leading-snug"
											>
												<p>Labels to add to items after processing.</p>
											</PopoverContent>
										</Popover>
									)}
								</div>
								<div className="flex flex-wrap items-center gap-2">
									{addLabels.length === 0 && (
										<span className="inline-flex h-7 items-center rounded-full border border-dashed px-3 text-[11px] text-muted-foreground">
											No labels to add
										</span>
									)}
									{addLabels.map((label, i) => (
										<span
											key={i}
											className={cn(
												"inline-flex items-center rounded-full px-3 py-1 text-[11px] font-normal cursor-pointer select-none transition duration-200",
												"bg-green-600 text-white",
												editing && "hover:bg-red-500 hover:text-white"
											)}
											onClick={() =>
												editing &&
												updateApplication(
													idx,
													"Add",
													addLabels.filter((_, j) => j !== i)
												)
											}
											title={editing ? "Remove label" : label}
										>
											{label}
										</span>
									))}
									{editing && (
										<form
											onSubmit={(e) => {
												e.preventDefault();
												const input = e.currentTarget.elements.namedItem(
													"addLabel"
												) as HTMLInputElement;
												const val = input.value.trim();
												if (val && !addLabels.includes(val)) {
													updateApplication(idx, "Add", [...addLabels, val]);
												}
												input.value = "";
											}}
											className="flex items-center gap-1"
										>
											<Input
												name="addLabel"
												placeholder="Add label..."
												className="h-7 w-40 text-xs"
											/>
											<Button
												type="submit"
												variant="outline"
												size="icon"
												className="h-7 w-7"
												title="Add label"
											>
												<Plus className="h-4 w-4" />
											</Button>
										</form>
									)}
								</div>
							</div>

							{/* Remove Labels */}
							<div className="space-y-2">
								<div className="flex items-center justify-between">
									<Label>Remove Labels</Label>
									{editing && (
										<Popover>
											<PopoverTrigger asChild>
												<Button
								variant="outline"
								className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
								aria-label="help-labels-tags-remove"
							>
								<HelpCircle className="h-4 w-4" />
							</Button>
											</PopoverTrigger>
											<PopoverContent
												side="right"
												align="center"
												sideOffset={8}
												className="w-72 text-xs leading-snug"
											>
												<p>Labels to remove from items after processing.</p>
											</PopoverContent>
										</Popover>
									)}
								</div>
								<div className="flex flex-wrap items-center gap-2">
									{removeLabels.length === 0 && (
										<span className="inline-flex h-7 items-center rounded-full border border-dashed px-3 text-[11px] text-muted-foreground">
											No labels to remove
										</span>
									)}
									{removeLabels.map((label, i) => (
										<span
											key={i}
											className={cn(
												"inline-flex items-center rounded-full px-3 py-1 text-[11px] font-normal cursor-pointer select-none transition duration-200",
												"bg-red-600 text-white",
												editing && "hover:bg-red-800 hover:text-white"
											)}
											onClick={() =>
												editing &&
												updateApplication(
													idx,
													"Remove",
													removeLabels.filter((_, j) => j !== i)
												)
											}
											title={editing ? "Remove label" : label}
										>
											{label}
										</span>
									))}
									{editing && (
										<form
											onSubmit={(e) => {
												e.preventDefault();
												const input = e.currentTarget.elements.namedItem(
													"removeLabel"
												) as HTMLInputElement;
												const val = input.value.trim();
												if (val && !removeLabels.includes(val)) {
													updateApplication(idx, "Remove", [...removeLabels, val]);
												}
												input.value = "";
											}}
											className="flex items-center gap-1"
										>
											<Input
												name="removeLabel"
												placeholder="Remove label..."
												className="h-7 w-40 text-xs"
											/>
											<Button
												type="submit"
												variant="outline"
												size="icon"
												className="h-7 w-7"
												title="Remove label"
											>
												<Plus className="h-4 w-4" />
											</Button>
										</form>
									)}
								</div>
							</div>

							{appError && <p className="text-xs text-red-500">{appError}</p>}
						</div>
					);
				})}

				{errors["Providers"] && <p className="text-xs text-red-500">{errors["Providers"]}</p>}
			</div>
		</Card>
	);
};
