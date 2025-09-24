"use client";

import { HelpCircle } from "lucide-react";

import React, { useEffect, useRef } from "react";

import Link from "next/link";

import { PopoverHelp } from "@/components/shared/popover-help";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import { AppConfigAutoDownload } from "@/types/config/config-app";

interface ConfigSectionAutoDownloadProps {
	value: AppConfigAutoDownload;
	editing: boolean;
	dirtyFields?: Partial<Record<keyof AppConfigAutoDownload, boolean>>;
	onChange: <K extends keyof AppConfigAutoDownload>(field: K, value: AppConfigAutoDownload[K]) => void;
	errorsUpdate?: (errors: Partial<Record<keyof AppConfigAutoDownload, string>>) => void;
}

const validateCron = (expr: string): string | null => {
	const trimmed = expr.trim();
	if (!trimmed) return "Cron expression is required when enabled.";
	const parts = trimmed.split(/\s+/);
	if (parts.length < 5 || parts.length > 6) return "Cron must have 5 or 6 fields.";
	// Light validation (allow common chars)
	if (!/^[\d*/,@\-\sA-Za-z]+$/.test(trimmed)) return "Cron contains invalid characters.";
	return null;
};

const parseCronToHumanReadable = (cronExpression: string): string => {
	const parts = cronExpression.split(" ");
	if (parts.length !== 5) return "Current cron expression is invalid";

	const [minute, hour, dayOfMonth, month, dayOfWeek] = parts;

	const minuteText = minute === "0" ? "at the start of" : `at minute ${minute}`;
	const hourText = hour === "*" ? "every hour" : `hour ${hour}`;
	const dayOfMonthText = dayOfMonth === "*" ? "every day" : `on day ${dayOfMonth}`;
	const monthText = month === "*" ? "every month" : `in month ${month}`;
	const dayOfWeekText = dayOfWeek === "*" ? "every day of the week" : `on day ${dayOfWeek} of the week`;

	return `Currently runs ${minuteText} ${hourText}, ${dayOfMonthText}, ${monthText}, and ${dayOfWeekText}.`;
};

export const ConfigSectionAutoDownload: React.FC<ConfigSectionAutoDownloadProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
}) => {
	const prevErrorsRef = useRef<string>("{}");
	const errors = React.useMemo(() => {
		const errs: Partial<Record<keyof AppConfigAutoDownload, string>> = {};
		if (value.Enabled) {
			const cronErr = validateCron(value.Cron);
			if (cronErr) errs.Cron = cronErr;
		}
		return errs;
	}, [value.Enabled, value.Cron]);

	useEffect(() => {
		if (!errorsUpdate) return;
		const ser = JSON.stringify(errors);
		if (ser === prevErrorsRef.current) return;
		prevErrorsRef.current = ser;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	return (
		<Card className="p-5 space-y-1">
			<h2 className="text-xl font-semibold">Auto Download</h2>

			{/* Enabled */}
			<div
				className={cn(
					"flex items-center justify-between border rounded-md p-3 transition",
					"border-muted",
					dirtyFields.Enabled && "border-amber-500"
				)}
			>
				<Label className="mr-2">Enabled</Label>
				<div className="flex items-center gap-2">
					<Switch
						disabled={!editing}
						checked={value.Enabled}
						onCheckedChange={(v) => onChange("Enabled", v)}
					/>
					{editing && (
						<PopoverHelp ariaLabel="help-auto-download-enabled">
							<p>
								Auto Download will check periodically for new updates your saved sets. This is helpful
								if you want to download and apply season poster/titlecard updates from future updates
								your sets.
							</p>
						</PopoverHelp>
					)}
				</div>
			</div>

			{/* Cron */}
			<div
				className={cn(
					"space-y-1",
					(value.Cron || errors.Cron || dirtyFields.Cron) && "rounded-md",
					errors.Cron ? "border border-red-500 p-3" : dirtyFields.Cron && "border border-amber-500 p-3"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>Cron Expression</Label>
					{editing && (
						<PopoverHelp ariaLabel="help-cron-expression">
							<p>
								Cron expression format. Use the standard cron syntax to specify when the job should run.
								You can use a site like{" "}
								<Link
									className="text-primary hover:underline"
									href="https://crontab.guru/"
									target="_blank"
									rel="noopener noreferrer"
								>
									crontab.guru
								</Link>{" "}
								to help you create and test your cron expressions.
							</p>
						</PopoverHelp>
					)}
				</div>
				<Input
					disabled={!editing || !value.Enabled}
					placeholder="e.g. 0 3 * * *"
					value={value.Cron}
					onChange={(e) => onChange("Cron", e.target.value)}
					aria-invalid={!!errors.Cron}
				/>
				{errors.Cron && <p className="text-xs text-red-500">{errors.Cron}</p>}
				{value.Enabled && !errors.Cron && value.Cron.trim() && (
					<p className="text-[10px] text-muted-foreground">{parseCronToHumanReadable(value.Cron)}</p>
				)}
			</div>
		</Card>
	);
};
