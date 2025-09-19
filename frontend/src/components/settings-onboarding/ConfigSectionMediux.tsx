"use client";

import { checkMediuxNewTokenStatusResult } from "@/services/settings-onboarding/api-mediux-connection";

import React, { useCallback, useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { cn } from "@/lib/cn";

import { AppConfigMediux } from "@/types/config/config-app";

interface ConfigSectionMediuxProps {
	value: AppConfigMediux;
	editing: boolean;
	configAlreadyLoaded: boolean;
	dirtyFields?: Partial<Record<keyof AppConfigMediux, boolean>>;
	onChange: <K extends keyof AppConfigMediux>(field: K, value: AppConfigMediux[K]) => void;
	errorsUpdate?: (errors: Partial<Record<keyof AppConfigMediux, string>>) => void;
}

const QUALITY_OPTIONS = ["optimized", "original"] as const;

export const ConfigSectionMediux: React.FC<ConfigSectionMediuxProps> = ({
	value,
	editing,
	configAlreadyLoaded,
	dirtyFields = {},
	onChange,
	errorsUpdate,
}) => {
	const prevErrorsRef = useRef<string>("");

	const [remoteTokenError, setRemoteTokenError] = useState<string | null>(null);
	const [testingToken, setTestingToken] = useState(false);

	// Validation (local + remote)
	const errors = React.useMemo(() => {
		const errs: Partial<Record<keyof AppConfigMediux, string>> = {};
		if (!value.Token.trim()) errs.Token = "Token is required.";
		if (!value.DownloadQuality.trim()) errs.DownloadQuality = "Select a download quality.";
		else if (!QUALITY_OPTIONS.includes(value.DownloadQuality as (typeof QUALITY_OPTIONS)[number]))
			errs.DownloadQuality = "Invalid quality option.";
		// Merge remote error (overrides local message for Token if present)
		if (remoteTokenError) errs.Token = remoteTokenError;
		return errs;
	}, [value.Token, value.DownloadQuality, remoteTokenError]);

	// Emit errors upward
	useEffect(() => {
		if (!errorsUpdate) return;
		const serialized = JSON.stringify(errors);
		if (serialized === prevErrorsRef.current) return;
		prevErrorsRef.current = serialized;
		errorsUpdate(errors);
	}, [errors, errorsUpdate]);

	// Reset remote error when token text changes
	useEffect(() => {
		setRemoteTokenError(null);
	}, [value.Token]);

	const runRemoteValidation = useCallback(async () => {
		if (!value.Token.trim()) {
			setRemoteTokenError("Token is required.");
			return;
		}
		setTestingToken(true);
		const { ok, message } = await checkMediuxNewTokenStatusResult(value);
		setTestingToken(false);
		if (ok) {
			setRemoteTokenError(null);
		} else {
			setRemoteTokenError(message || "Token invalid");
		}
	}, [value, setRemoteTokenError, setTestingToken]);

	// If the config is already loaded, we can run validation
	useEffect(() => {
		if (configAlreadyLoaded) {
			runRemoteValidation();
		}
	}, [configAlreadyLoaded, runRemoteValidation]);

	return (
		<Card className="p-5 space-y-1">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">MediUX</h2>
				<Button
					variant="outline"
					size="sm"
					hidden={editing}
					disabled={editing || testingToken}
					onClick={() => {
						runRemoteValidation();
					}}
				>
					{testingToken ? "Testing..." : "Test Token"}
				</Button>
			</div>
			{/* Token */}
			<div
				className={cn(
					"space-y-1 border rounded-md p-3 transition",
					errors.Token ? "border-red-500" : dirtyFields.Token ? "border-amber-500" : "border-muted"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>Token</Label>
					{editing && (
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-mediux-token"
								>
									?
								</Button>
							</PopoverTrigger>
							<PopoverContent
								side="right"
								align="center"
								sideOffset={8}
								className="w-64 text-xs leading-snug"
							>
								<p>Mediux API token. Paste the personal token generated from your Mediux account.</p>
							</PopoverContent>
						</Popover>
					)}
				</div>
				<Input
					disabled={!editing}
					placeholder="Mediux API token"
					value={value.Token}
					onChange={(e) => onChange("Token", e.target.value)}
					aria-invalid={!!errors.Token}
					onBlur={() => {
						runRemoteValidation();
					}}
				/>
				{errors.Token && <p className="text-xs text-red-500">{errors.Token}</p>}
			</div>
			{/* Download Quality */}
			<div
				className={cn(
					"space-y-1 border rounded-md p-3 transition",
					errors.DownloadQuality
						? "border-red-500"
						: dirtyFields.DownloadQuality
							? "border-amber-500"
							: "border-muted"
				)}
			>
				<div className="flex items-center justify-between">
					<Label>Download Quality</Label>
					{editing && (
						<Popover>
							<PopoverTrigger asChild>
								<Button
									variant="outline"
									className="h-6 w-6 rounded-md border flex items-center justify-center text-xs bg-background hover:bg-muted transition"
									aria-label="help-mediux-download-quality"
								>
									?
								</Button>
							</PopoverTrigger>
							<PopoverContent
								side="right"
								align="center"
								sideOffset={8}
								className="w-64 text-xs leading-snug"
							>
								<p className="mb-2">Select the desired quality for fetched downloads:</p>
								<ul className="space-y-1">
									<li className="flex gap-2">
										<span className="inline-flex h-5 items-center rounded-sm bg-muted px-2 font-mono text-[10px]">
											optimized
										</span>
										<span className="text-[11px]">Balanced for size & performance.</span>
									</li>
									<li className="flex gap-2">
										<span className="inline-flex h-5 items-center rounded-sm bg-muted px-2 font-mono text-[10px]">
											original
										</span>
										<span className="text-[11px]">Full source quality (largest size).</span>
									</li>
								</ul>
							</PopoverContent>
						</Popover>
					)}
				</div>
				<Select
					disabled={!editing}
					value={value.DownloadQuality}
					onValueChange={(v) => onChange("DownloadQuality", v as AppConfigMediux["DownloadQuality"])}
				>
					<SelectTrigger className="w-full" id="mediux-download-quality-trigger">
						<SelectValue placeholder="Select quality..." />
					</SelectTrigger>
					<SelectContent>
						{QUALITY_OPTIONS.map((q) => (
							<SelectItem key={q} value={q}>
								{q}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
				{errors.DownloadQuality && <p className="text-xs text-red-500">{errors.DownloadQuality}</p>}
			</div>
		</Card>
	);
};
