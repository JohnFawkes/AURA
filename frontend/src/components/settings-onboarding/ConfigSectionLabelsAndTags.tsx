"use client";

import { Plus, Trash2 } from "lucide-react";

import React, { useEffect, useRef, useState } from "react";

import { PopoverHelp } from "@/components/shared/popover-help";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
	mediaServerType?: string;
	srOptions?: string[];
}

export const ConfigSectionLabelsAndTags: React.FC<ConfigSectionLabelsAndTagsProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
	mediaServerType,
	srOptions,
}) => {
	const prevErrorsRef = useRef<string>("");

	const APPLICATION_TYPES = React.useMemo(() => {
		let appTypes = [];
		if (mediaServerType === "Plex") {
			appTypes.push("Plex");
		}
		if (Array.isArray(srOptions) && srOptions.length > 0) {
			appTypes = appTypes.concat(srOptions);
		}
		return appTypes;
	}, [mediaServerType, srOptions]);

	const [newApplicationType, setNewApplicationType] = useState(APPLICATION_TYPES[0] || "");

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
			const orLabelTag = app.Application === "Plex" ? "Label" : "Tag";
			if (app.Enabled && addLabels.length === 0 && removeLabels.length === 0) {
				errs[`Applications.[${idx}]`] =
					`At least one ${orLabelTag.toLowerCase()} to add or remove must be specified.`;
			}
			const intersection = addLabels.filter((label) => removeLabels.includes(label));
			if (intersection.length > 0) {
				errs[`Applications.[${idx}]`] =
					`${orLabelTag} cannot be in both Add and Remove: ${intersection.join(", ")}`;
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
		} else if (Array.isArray(srOptions) && srOptions.includes(type)) {
			newEntry = { Application: type, Enabled: true, Add: [], Remove: [] };
		}
		setApplications([...applications, newEntry!]);
	};

	useEffect(() => {
		const usedTypes = applications.map((app) => app.Application);
		const nextAvailable = APPLICATION_TYPES.find((type) => !usedTypes.includes(type)) || "";
		setNewApplicationType(nextAvailable);
	}, [APPLICATION_TYPES, applications]);

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
		<Card className={`p-5 ${Object.values(errors).some(Boolean) ? "border-red-500" : "border-muted"}`}>
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold text-blue-500">Labels & Tags</h2>
			</div>
			<div className={cn("space-y-4", "rounded-md")}>
				<div className="flex items-center justify-between">
					{editing &&
						(() => {
							const availableTypes = APPLICATION_TYPES.filter(
								(p) => !applications.some((app) => app.Application === p)
							);
							return availableTypes.length > 0 ? (
								<div className="flex items-center gap-2">
									<Select value={newApplicationType} onValueChange={(v) => setNewApplicationType(v)}>
										<SelectTrigger className="h-8 w-36">
											<SelectValue placeholder="Type" />
										</SelectTrigger>
										<SelectContent>
											{availableTypes.map((p) => (
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
							) : null;
						})()}
				</div>

				{!mediaServerType && (!srOptions || srOptions.length === 0) ? (
					<p className="text-sm text-red-500 mb-2">
						You need to have a Media Server configured and/or at least one Sonarr/Radarr instance before you
						can use Labels &amp; Tags.
					</p>
				) : applications.length === 0 ? (
					<p className="text-xs text-muted-foreground">No applications configured.</p>
				) : null}

				{applications.map((app, idx) => {
					const appDirty = dirtyFields.Applications?.[idx] as Partial<{
						Enabled?: boolean;
						Add?: boolean;
						Remove?: boolean;
					}>;
					const appError = errors[`Applications.[${idx}]`];
					const addLabels = Array.isArray(app.Add) ? app.Add : [];
					const removeLabels = Array.isArray(app.Remove) ? app.Remove : [];
					const orLabelTag = app.Application === "Plex" ? "Labels" : "Tags";
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
									<p className="font-medium text-lg">{app.Application}</p>
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
										//disabled={applications.length === 1}
										className="bg-red-700"
									>
										<Trash2 className="h-4 w-4" />
									</Button>
								)}
							</div>

							{/* Add Labels/Tags */}
							<div className="space-y-2">
								<div className="flex items-center justify-between">
									<Label>Add {orLabelTag}</Label>
									{editing && (
										<PopoverHelp ariaLabel="help-labels-to-add">
											<p>{orLabelTag} to add to items after processing.</p>
										</PopoverHelp>
									)}
								</div>
								<div className="flex flex-wrap items-center gap-2">
									{addLabels.length === 0 && (
										<span className="inline-flex h-7 items-center rounded-full border border-dashed px-3 text-xs text-muted-foreground">
											No {orLabelTag.toLowerCase()} to add
										</span>
									)}
									{addLabels.map((label, i) => (
										<span
											key={i}
											className={cn(
												"inline-flex items-center rounded-full px-3 py-1 text-xs font-normal cursor-pointer select-none transition duration-200",
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
											title={editing ? `Remove ${orLabelTag}` : label}
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
												placeholder={`Add ${orLabelTag}...`}
												className="h-7 w-40 text-xs"
											/>
											<Button
												type="submit"
												variant="outline"
												size="icon"
												className="h-7 w-7"
												title={`Add ${orLabelTag}`}
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
									<Label>Remove {orLabelTag}</Label>
									{editing && (
										<PopoverHelp ariaLabel="help-labels-tags-remove">
											<p>{orLabelTag} to remove from items after processing.</p>
										</PopoverHelp>
									)}
								</div>
								<div className="flex flex-wrap items-center gap-2">
									{removeLabels.length === 0 && (
										<span className="inline-flex h-7 items-center rounded-full border border-dashed px-3 text-xs text-muted-foreground">
											No {orLabelTag.toLowerCase()} to remove
										</span>
									)}
									{removeLabels.map((label, i) => (
										<span
											key={i}
											className={cn(
												"inline-flex items-center rounded-full px-3 py-1 text-xs font-normal cursor-pointer select-none transition duration-200",
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
											title={editing ? `Remove ${orLabelTag}` : label}
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
												placeholder={`Remove ${orLabelTag}...`}
												className="h-7 w-40 text-xs"
											/>
											<Button
												type="submit"
												variant="outline"
												size="icon"
												className="h-7 w-7"
												title={`Remove ${orLabelTag}`}
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
			</div>
		</Card>
	);
};
