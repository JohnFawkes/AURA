"use client";

import { checkMediuxNewTokenStatusResult } from "@/services/settings-onboarding/api-mediux-connection";

import React, { useCallback, useEffect, useRef, useState } from "react";

import { GetConnectionColor } from "@/components/settings-onboarding/ConfigSectionMediaServer";
import {
	CONNECTION_STATUS_COLORS_BG,
	ConfigConnectionStatus,
} from "@/components/settings-onboarding/ConfigSectionSonarrRadarr";
import { PopoverHelp } from "@/components/shared/popover-help";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
	const [connectionStatus, setConnectionStatus] = useState<ConfigConnectionStatus>({
		status: "unknown",
		color: GetConnectionColor("unknown"),
	});

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

	const runRemoteValidation = useCallback(
		async (showToast = true) => {
			if (!value.Token.trim()) {
				setRemoteTokenError("Token is required.");
				setConnectionStatus({ status: "error", color: GetConnectionColor("error") });
				return;
			}

			setTestingToken(true);
			const start = Date.now();
			const { ok, message } = await checkMediuxNewTokenStatusResult(value, showToast);
			const elapsed = Date.now() - start;
			const minDelay = 400; // milliseconds

			if (elapsed < minDelay) {
				await new Promise((resolve) => setTimeout(resolve, minDelay - elapsed));
			}

			setTestingToken(false);

			if (ok) {
				setRemoteTokenError(null);
				setConnectionStatus({ status: "ok", color: GetConnectionColor("ok") });
			} else {
				setRemoteTokenError(message || "Token invalid");
				setConnectionStatus({ status: "error", color: GetConnectionColor("error") });
			}
		},
		[value, setRemoteTokenError, setTestingToken]
	);

	// If the config is already loaded, we can run validation
	useEffect(() => {
		if (configAlreadyLoaded) {
			runRemoteValidation(false);
		}
	}, [configAlreadyLoaded, runRemoteValidation]);

	return (
		<Card className={`p-5 ${Object.values(errors).some(Boolean) ? "border-red-500" : "border-muted"}`}>
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-2">
					<h2 className={`text-xl font-semibold text-${connectionStatus.color}`}>MediUX</h2>
					<span
						className={`h-2 w-2 rounded-full ${CONNECTION_STATUS_COLORS_BG[connectionStatus.status]} animate-pulse`}
						title={`Connection status: ${connectionStatus.status}`}
					/>
				</div>
				<Button
					variant="outline"
					size="sm"
					hidden={editing}
					disabled={editing || testingToken}
					onClick={() => runRemoteValidation()}
					className="cursor-pointer hover:text-primary"
				>
					{testingToken ? "Testing..." : "Test Token"}
				</Button>
			</div>
			{/* Token */}
			<div className={cn("space-y-1 border rounded-md p-3 transition")}>
				<div className="flex items-center justify-between">
					<Label>Token</Label>
					{editing && (
						<PopoverHelp ariaLabel="help-mediux-token">
							<p>Mediux API token. Paste the personal token generated from your Mediux account.</p>
						</PopoverHelp>
					)}
				</div>
				<Input
					disabled={!editing}
					placeholder="Mediux API token"
					value={value.Token}
					onChange={(e) => onChange("Token", e.target.value)}
					onBlur={() => {
						runRemoteValidation();
					}}
					className={cn(dirtyFields.Token && "border-amber-500")}
				/>
				{errors.Token && <p className="text-xs text-red-500">{errors.Token}</p>}
			</div>

			{/* Download Quality */}
			<div className="space-y-1 border rounded-md p-3 transition">
				<div className="flex items-center justify-between">
					<Label>Download Quality</Label>
					{editing && (
						<PopoverHelp ariaLabel="help-mediux-download-quality">
							<p className="mb-2">Select the desired quality for fetched downloads:</p>
							<ul className="space-y-2">
								<li>
									<div className="flex items-center gap-3 rounded-md border border-muted bg-muted/50 px-3 py-2">
										<p className="inline-flex items-center rounded bg-background px-2 py-1 font-mono">
											optimized
										</p>
										<p>Balanced for size & performance.</p>
									</div>
								</li>
								<li>
									<div className="flex items-center gap-3 rounded-md border border-muted bg-muted/50 px-3 py-2">
										<p className="inline-flex items-center rounded bg-background px-2 py-1 font-mono">
											original
										</p>
										<p>Full source quality (largest size).</p>
									</div>
								</li>
							</ul>
						</PopoverHelp>
					)}
				</div>
				<Select
					disabled={!editing}
					value={value.DownloadQuality}
					onValueChange={(v) => onChange("DownloadQuality", v as AppConfigMediux["DownloadQuality"])}
				>
					<SelectTrigger
						id="mediux-download-quality-trigger"
						className={cn("w-full", dirtyFields.DownloadQuality && "border-amber-500")}
					>
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
