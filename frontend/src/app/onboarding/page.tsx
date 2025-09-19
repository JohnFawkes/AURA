"use client";

import { updateConfig } from "@/services/settings-onboarding/api-config-update";
import { finalizeOnboarding } from "@/services/settings-onboarding/api-onboarding-finalize";
import yaml from "js-yaml";
import { toast } from "sonner";

import { JSX, useCallback, useEffect, useMemo, useState } from "react";

import Image from "next/image";

import { ConfigSectionAuth } from "@/components/settings-onboarding/ConfigSectionAuth";
import { ConfigSectionAutoDownload } from "@/components/settings-onboarding/ConfigSectionAutoDownload";
import { ConfigSectionImages } from "@/components/settings-onboarding/ConfigSectionImages";
import { ConfigSectionKometa } from "@/components/settings-onboarding/ConfigSectionKometa";
import { ConfigSectionLogging } from "@/components/settings-onboarding/ConfigSectionLogging";
import { ConfigSectionMediaServer } from "@/components/settings-onboarding/ConfigSectionMediaServer";
import { ConfigSectionMediux } from "@/components/settings-onboarding/ConfigSectionMediux";
import { ConfigSectionNotifications } from "@/components/settings-onboarding/ConfigSectionNotifications";
import { Button } from "@/components/ui/button";
import { H1, H2, P } from "@/components/ui/typography";

import { useOnboardingStore } from "@/lib/stores/global-store-onboarding";

import type { AppConfig } from "@/types/config/config-app";
import { defaultAppConfig } from "@/types/config/config-default-app";

interface StepDef {
	key: string;
	title: string;
	optional?: boolean;
	render: () => JSX.Element;
}

const OnboardingPage = () => {
	const { status, fetchStatus } = useOnboardingStore();

	// Hydrate/fetch onboarding status on mount
	useEffect(() => {
		if (!status) fetchStatus();
	}, [status, fetchStatus]);

	const [applyLoading, setApplyLoading] = useState(false);
	const [configState, setConfigState] = useState<AppConfig>(() => status?.currentSetup || defaultAppConfig());
	const [validationErrors, setValidationErrors] = useState<Record<string, Record<string, string>>>({});
	const [errorSummaryOpen, setErrorSummaryOpen] = useState(false);

	// Keep configState in sync with backend status if it changes
	useEffect(() => {
		if (status?.currentSetup) {
			setConfigState(status.currentSetup);
		}
	}, [status?.currentSetup]);

	const updateSectionErrors = useCallback((section: string, errs?: Record<string, string>) => {
		setValidationErrors((prev) => {
			if (!errs || Object.keys(errs).length === 0) {
				// eslint-disable-next-line @typescript-eslint/no-unused-vars
				const { [section]: _, ...rest } = prev;
				return rest;
			}
			return { ...prev, [section]: errs };
		});
	}, []);

	const updateSectionField = useCallback(
		<S extends keyof AppConfig, K extends keyof AppConfig[S]>(section: S, field: K, value: AppConfig[S][K]) => {
			setConfigState(
				(prev) =>
					({
						...prev,
						[section]: { ...(prev[section] as object), [field]: value },
					}) as AppConfig
			);
		},
		[]
	);

	const updateImagesField = useCallback(
		<G extends keyof AppConfig["Images"], F extends keyof AppConfig["Images"][G]>(
			group: G,
			field: F,
			value: AppConfig["Images"][G][F]
		) => {
			setConfigState((prev) => ({
				...prev,
				Images: {
					...prev.Images,
					[group]: {
						...prev.Images[group],
						[field]: value,
					},
				},
			}));
		},
		[]
	);

	// Memoized YAML representation
	const reviewYaml = useMemo(() => {
		try {
			// Clone to strip proxies / undefined
			const plain = JSON.parse(JSON.stringify(configState));
			return yaml.dump(plain, {
				noRefs: true, // avoid anchors
				lineWidth: 100, // prevent excessive wrapping
				skipInvalid: true,
			});
		} catch {
			return "# Failed to serialize configuration to YAML";
		}
	}, [configState]);

	const steps: StepDef[] = useMemo(
		() => [
			{
				key: "welcome",
				title: "Welcome",
				render: () => (
					<div className="space-y-6">
						<div className="flex items-center gap-4">
							<H2 className="text-4xl font-bold">Welcome to</H2>
							<Image src="/aura_word_logo.svg" alt="Aura Logo" width={120} height={120} />
						</div>

						<P className="text-muted-foreground max-w-xl">
							{status?.configLoaded && (
								<span className="text-destructive">
									Your configuration file might have some errors.{" "}
								</span>
							)}
							This quick setup will guide you through core configuration. Use Next / Back to move through
							the steps.
						</P>
					</div>
				),
			},
			{
				key: "mediux",
				title: "Mediux",
				render: () => (
					<ConfigSectionMediux
						value={configState.Mediux}
						editing
						configAlreadyLoaded={status?.configLoaded || false}
						dirtyFields={{}}
						onChange={(f, v) => updateSectionField("Mediux", f, v)}
						errorsUpdate={(errs) => updateSectionErrors("Mediux", errs as Record<string, string>)}
					/>
				),
			},
			{
				key: "mediaserver",
				title: "Media Server",
				render: () => (
					<ConfigSectionMediaServer
						value={configState.MediaServer}
						editing
						configAlreadyLoaded={status?.configLoaded || false}
						dirtyFields={{}}
						onChange={(f, v) => updateSectionField("MediaServer", f, v)}
						errorsUpdate={(errs) => updateSectionErrors("MediaServer", errs as Record<string, string>)}
					/>
				),
			},
			{
				key: "auth",
				title: "Auth",
				optional: true,
				render: () => (
					<ConfigSectionAuth
						value={configState.Auth}
						editing
						dirtyFields={{}}
						onChange={(f, v) => updateSectionField("Auth", f, v)}
						errorsUpdate={(errs) => updateSectionErrors("Auth", errs as Record<string, string>)}
					/>
				),
			},
			{
				key: "logging",
				title: "Logging",
				render: () => (
					<ConfigSectionLogging
						value={configState.Logging}
						editing
						dirtyFields={{}}
						onChange={(f, v) => updateSectionField("Logging", f, v)}
						errorsUpdate={(errs) => updateSectionErrors("Logging", errs as Record<string, string>)}
					/>
				),
			},
			{
				key: "images",
				title: "Images",
				optional: true,
				render: () => (
					<ConfigSectionImages
						value={configState.Images}
						editing
						dirtyFields={{}}
						onChange={updateImagesField}
						errorsUpdate={(errs) => updateSectionErrors("Images", errs as Record<string, string>)}
					/>
				),
			},
			{
				key: "kometa",
				title: "Kometa",
				optional: true,
				render: () => (
					<ConfigSectionKometa
						value={configState.Kometa}
						editing
						dirtyFields={{}}
						onChange={(f, v) => updateSectionField("Kometa", f, v)}
						errorsUpdate={(errs) => updateSectionErrors("Kometa", errs as Record<string, string>)}
					/>
				),
			},
			{
				key: "notifications",
				title: "Notifications",
				optional: true,
				render: () => (
					<ConfigSectionNotifications
						value={configState.Notifications}
						editing
						dirtyFields={{}}
						onChange={(f, v) => updateSectionField("Notifications", f, v)}
						errorsUpdate={(errs) => updateSectionErrors("Notifications", errs as Record<string, string>)}
					/>
				),
			},
			{
				key: "autodownload",
				title: "Auto Download",
				optional: true,
				render: () => (
					<ConfigSectionAutoDownload
						value={configState.AutoDownload}
						editing
						dirtyFields={{}}
						onChange={(f, v) => updateSectionField("AutoDownload", f, v)}
						errorsUpdate={(errs) => updateSectionErrors("AutoDownload", errs as Record<string, string>)}
					/>
				),
			},
			{
				key: "review",
				title: "Review & Finish",
				render: () => (
					<div className="space-y-6">
						<h2 className="text-2xl font-semibold">Review</h2>
						<p className="text-sm text-muted-foreground">
							Press Finish to apply configuration and start the application.
						</p>
						<pre className="p-4 bg-muted rounded text-xs overflow-auto max-h-96">{reviewYaml}</pre>
						{Object.keys(validationErrors).length > 0 && (
							<div className="text-red-500 text-sm">Resolve validation errors before finishing.</div>
						)}
					</div>
				),
			},
		],
		[
			status?.configLoaded,
			configState.Mediux,
			configState.MediaServer,
			configState.Auth,
			configState.Logging,
			configState.Images,
			configState.Kometa,
			configState.Notifications,
			configState.AutoDownload,
			updateSectionField,
			updateSectionErrors,
			updateImagesField,
			reviewYaml,
			validationErrors,
		]
	);

	const [index, setIndex] = useState(0);
	const current = steps[index];
	const lastIndex = steps.length - 1;
	const hasErrors = Object.keys(validationErrors).length > 0;

	const next = () => setIndex((i) => Math.min(i + 1, lastIndex));
	const prev = () => setIndex((i) => Math.max(i - 1, 0));
	const skipOptional = () => {
		for (let i = index + 1; i <= lastIndex; i++) {
			if (!steps[i].optional) {
				setIndex(i);
				return;
			}
		}
		setIndex(lastIndex);
	};

	const finish = async () => {
		if (hasErrors) {
			toast.error("Fix validation errors first.");
			return;
		}
		setApplyLoading(true);

		try {
			const resp = await updateConfig(configState);
			if (resp.status === "success") {
				if (resp.data) {
					const finalizeResp = await finalizeOnboarding(resp.data);
					if (finalizeResp.status === "success") {
						toast.success("Configuration applied successfully, redirecting...");
						setTimeout(() => (window.location.href = "/"), 50);
					} else {
						toast.error("Failed to finalize onboarding.");
					}
				} else {
					toast.error("Failed to apply configuration.");
				}
			} else if (resp.status === "warn") {
				toast.warning(
					typeof resp.data === "string" ? resp.data : "Warning occurred while applying configuration."
				);
			} else {
				toast.error("Failed to apply configuration.");
			}
		} catch {
			toast.error("An error occurred while applying configuration.");
		}

		setApplyLoading(false);
	};

	// Build a lookup from (lowercased) step key to its index for quick jumps
	const stepIndexByKey = useMemo(() => {
		const m: Record<string, number> = {};
		steps.forEach((s, i) => {
			m[s.key.toLowerCase()] = i;
		});
		return m;
	}, [steps]);

	// Map config section names (e.g. MediaServer) to step keys (lowercase)
	const jumpToSection = (sectionName: string) => {
		const target = stepIndexByKey[sectionName.toLowerCase()];
		if (typeof target === "number") {
			setIndex(target);
			requestAnimationFrame(() => {
				window.scrollTo({ top: 0, behavior: "smooth" });
			});
		}
	};

	const ErrorSummary = ({ errors }: { errors: Record<string, Record<string, string>> }) => {
		const sections = Object.entries(errors);
		const total = sections.reduce((sum, [, errs]) => sum + Object.keys(errs).length, 0);
		if (total === 0) return null;

		return (
			<div className="w-full rounded-md border border-destructive/30 bg-destructive/5 p-3 mt-2">
				<div className="flex items-center justify-between gap-4">
					<P className="m-0 text-sm font-medium text-destructive">
						{total} validation error{total > 1 ? "s" : ""}
					</P>
					<Button
						type="button"
						variant="ghost"
						size="sm"
						onClick={() => setErrorSummaryOpen((o) => !o)}
						className="h-6 px-2"
					>
						{errorSummaryOpen ? "Hide" : "Show"}
					</Button>
				</div>

				{errorSummaryOpen && (
					<div className="mt-3 grid gap-2 sm:grid-cols-2">
						{sections.map(([section, errs]) => {
							const count = Object.keys(errs).length;
							return (
								<div
									key={section}
									role="button"
									tabIndex={0}
									onClick={() => jumpToSection(section)}
									onKeyDown={(e) => {
										if (e.key === "Enter" || e.key === " ") {
											e.preventDefault();
											jumpToSection(section);
										}
									}}
									className="group cursor-pointer rounded border border-destructive/30 bg-destructive/10 p-2 transition hover:border-destructive/60 hover:bg-destructive/15 focus:outline-none focus:ring-2 focus:ring-destructive/60"
								>
									<div className="flex items-center justify-between">
										<p className="m-0 font-semibold text-sm text-destructive group-hover:underline">
											{section.replace(/([A-Z])/g, " $1")}
										</p>
										<span className="rounded bg-destructive/20 px-2 py-0.5 text-sm text-destructive">
											{count} error{count > 1 ? "s" : ""}
										</span>
									</div>
									<ul className="mt-1 space-y-0.5">
										{Object.entries(errs).map(([field, msg]) => (
											<li key={field} className="text-sm text-destructive">
												<span className="font-mono">{field}</span>
												<span className="mx-1 opacity-60"> - </span>
												{msg}
											</li>
										))}
									</ul>
								</div>
							);
						})}
					</div>
				)}
			</div>
		);
	};

	return (
		<div className="mx-auto max-w-5xl p-6 space-y-8">
			<div>
				<H1 className="text-2xl font-bold">Onboarding</H1>
			</div>
			<div className="flex items-center justify-between">
				<P className="mb-2 text-md text-muted-foreground">
					Step {index + 1} of {steps.length}: {current.title}
				</P>
				<div className="mb-2 flex flex-row flex-wrap gap-2 justify-end">
					{index > 0 && (
						<Button variant="outline" onClick={prev} disabled={applyLoading}>
							← Back
						</Button>
					)}
					{current.optional && index < lastIndex - 1 && (
						<Button variant="secondary" onClick={skipOptional} disabled={applyLoading}>
							Skip
						</Button>
					)}
					{index < lastIndex && (
						<Button onClick={next} disabled={applyLoading}>
							Next →
						</Button>
					)}
					{index === lastIndex && (
						<Button onClick={finish} disabled={applyLoading || hasErrors}>
							{applyLoading ? "Applying..." : "Apply & Save"}
						</Button>
					)}
				</div>
			</div>

			<div className="border rounded-lg p-5 bg-background shadow-sm">{current.render()}</div>

			{/* Nav + Optional label */}
			<div className="flex flex-row flex-wrap items-center justify-between gap-3 mt-2">
				<div className="text-xs text-muted-foreground min-h-[1rem]">
					{current.optional && <P className="text-sm font-medium text-amber-600 m-0">Optional step</P>}
				</div>
			</div>

			{/* Error summary below full width */}
			{hasErrors && <ErrorSummary errors={validationErrors} />}
		</div>
	);
};

export default OnboardingPage;
