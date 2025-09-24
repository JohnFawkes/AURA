"use client";

import { postClearOldLogs } from "@/services/settings-onboarding/api-logs-actions";
import { toast } from "sonner";

import React, { useEffect, useRef } from "react";

import { useRouter } from "next/navigation";

import { PopoverHelp } from "@/components/shared/popover-help";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { cn } from "@/lib/cn";

import { AppConfigLogging } from "@/types/config/config-app";

interface ConfigSectionLoggingProps {
	value: AppConfigLogging;
	editing: boolean;
	dirtyFields?: Partial<Record<keyof AppConfigLogging, boolean>>;
	onChange: <K extends keyof AppConfigLogging>(field: K, value: AppConfigLogging[K]) => void;
	errorsUpdate?: (errors: Partial<Record<keyof AppConfigLogging, string>>) => void;
}

const LOG_LEVEL_OPTIONS = ["TRACE", "DEBUG", "INFO", "WARN", "ERROR"] as const;

export const ConfigSectionLogging: React.FC<ConfigSectionLoggingProps> = ({
	value,
	editing,
	dirtyFields = {},
	onChange,
	errorsUpdate,
}) => {
	const router = useRouter();
	const prevErrorsRef = useRef<string>("");

	const errors = React.useMemo(() => {
		const errs: Partial<Record<keyof AppConfigLogging, string>> = {};
		if (!value.Level?.trim()) errs.Level = "Level is required.";
		else if (!LOG_LEVEL_OPTIONS.includes(value.Level as (typeof LOG_LEVEL_OPTIONS)[number]))
			errs.Level = "Invalid log level.";
		return errs;
	}, [value.Level]);

	useEffect(() => {
		if (!errorsUpdate) return;
		const ser = JSON.stringify(errors);
		if (ser === prevErrorsRef.current) return;
		prevErrorsRef.current = ser;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	const handleViewLogs = () => {
		router.push("/logs");
	};

	const handleClearOldLogs = async () => {
		try {
			const response = await postClearOldLogs();
			if (response.status === "error") {
				toast.error(response.error?.Message || "Failed to clear old logs");
				return;
			}
			toast.success(response.data || "Old logs cleared");
		} catch {
			toast.error("An unexpected error occurred");
		}
	};

	return (
		<Card className="p-5 space-y-1">
			<h2 className="text-xl font-semibold">Logging</h2>

			{/* Level */}
			<div
				className={cn(
					"space-y-1 border rounded-md p-3 transition",
					errors.Level ? "border-red-500" : dirtyFields.Level ? "border-amber-500" : "border-muted"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>Level</Label>
					{editing && (
						<PopoverHelp ariaLabel="help-logging-level">
							<p className="mb-2">Select verbosity of application logs.</p>
							<ul className="space-y-2 text-xs">
								<li>
									<span className="font-mono">TRACE</span> – detailed tracing information
								</li>
								<li>
									<span className="font-mono">DEBUG</span> – detailed development info
								</li>
								<li>
									<span className="font-mono">INFO</span> – high-level operational events
								</li>
								<li>
									<span className="font-mono">WARN</span> – unexpected but non-fatal issues
								</li>
								<li>
									<span className="font-mono">ERROR</span> – failures requiring attention
								</li>
							</ul>
						</PopoverHelp>
					)}
				</div>
				<Select
					disabled={!editing}
					value={value.Level}
					onValueChange={(v) => onChange("Level", v as AppConfigLogging["Level"])}
				>
					<SelectTrigger className="w-full" id="logging-level-trigger">
						<SelectValue placeholder="Select level..." />
					</SelectTrigger>
					<SelectContent>
						{LOG_LEVEL_OPTIONS.map((lvl) => (
							<SelectItem key={lvl} value={lvl}>
								{lvl}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
				{errors.Level && <p className="text-xs text-red-500">{errors.Level}</p>}
			</div>

			{/* File (read-only) */}
			<div className="space-y-1 border rounded-md p-3">
				<div className="flex items-center justify-between mb-1">
					<Label>Log File</Label>
					<div className="flex items-center gap-2">
						<Button hidden={editing} variant="outline" onClick={handleViewLogs}>
							View
						</Button>
						<Button hidden={editing} variant="destructive" onClick={handleClearOldLogs}>
							Clear Old
						</Button>
					</div>
				</div>
				<Input disabled value={value.File || ""} placeholder="App Log File Path" />
			</div>
		</Card>
	);
};
