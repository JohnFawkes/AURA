"use client";

import { useEffect, useState } from "react";
import { FaDiscord, FaGithub } from "react-icons/fa";

import Image from "next/image";
import Link from "next/link";

import { ReleaseNotesDialog } from "@/components/layout/app-release-notes";

import { log } from "@/lib/logger";

interface AppFooterProps {
	version?: string;
}

export function AppFooter({ version = "dev" }: AppFooterProps) {
	// Get the latest version from Github
	const [latestVersion, setLatestVersion] = useState<string | null>(null);

	const [showReleaseNotes, setShowReleaseNotes] = useState(false);
	const [changelog, setChangelog] = useState("");

	function isNewerVersion(latest: string, current: string): boolean {
		const parse = (v: string) => v.replace(/^v/, "").split("-")[0].split(".").map(Number);
		const [lMaj, lMin, lPatch] = parse(latest);
		const [cMaj, cMin, cPatch] = parse(current);

		if (lMaj > cMaj) return true;
		if (lMaj < cMaj) return false;
		if (lMin > cMin) return true;
		if (lMin < cMin) return false;
		if (lPatch > cPatch) return true;
		return false;
	}

	function normalizeVersion(version: string) {
		return version.replace(/^v/, "").replace(/-.*$/, "");
	}

	function getChangelogEntriesSince(changelog: string, lastVersion: string) {
		const regex = /## \[([^\]]+)\] - (\d{4}-\d{2}-\d{2})/g;
		const entries: { version: string; content: string }[] = [];
		let match;
		const indices: number[] = [];
		const versions: string[] = [];
		while ((match = regex.exec(changelog)) !== null) {
			indices.push(match.index);
			versions.push(match[1]);
		}
		indices.push(changelog.length);

		for (let i = 0; i < versions.length; i++) {
			entries.push({
				version: versions[i],
				content: changelog.slice(indices[i], indices[i + 1]),
			});
		}

		// Normalize lastSeenVersion for matching
		const normalizedLastVersion = normalizeVersion(lastVersion || "");
		const idx = entries.findIndex((e) => normalizeVersion(e.version) === normalizedLastVersion);

		return idx === -1 ? entries : entries.slice(0, idx);
	}

	useEffect(() => {
		const lastSeen = localStorage.getItem("lastSeenVersion");
		log("INFO", "App Footer", "Release Notes", `Last seen version: ${lastSeen}, Current version: ${version}`);

		fetch("/CHANGELOG.md")
			.then((res) => res.text())
			.then((fullChangelog) => {
				const relevantEntries = getChangelogEntriesSince(fullChangelog, lastSeen || "");
				// Join relevant entries for markdown rendering
				setChangelog(relevantEntries.map((e) => e.content).join("\n"));
				if (lastSeen !== version && relevantEntries.length > 0) {
					setShowReleaseNotes(true);
				}
			});
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [version]);

	function handleCloseReleaseNotes() {
		localStorage.setItem("lastSeenVersion", version);
		log("INFO", "App Footer", "Release Notes", `User closed release notes for version ${version}`);
		setShowReleaseNotes(false);
	}

	useEffect(() => {
		const fetchLatestVersion = async () => {
			try {
				const res = await fetch("https://raw.githubusercontent.com/mediux-team/AURA/master/VERSION.txt", {
					cache: "no-store",
				});
				if (!res.ok) throw new Error(`HTTP ${res.status}`);
				const txt = (await res.text()).trim();
				// Basic sanity (optional)
				if (txt && txt.length < 50) {
					setLatestVersion(txt);
				}
			} catch (error) {
				log("ERROR", "App Footer", "Fetch Latest Version", "Failed to fetch latest version:", error);
			}
		};
		fetchLatestVersion();
	}, []);

	return (
		<footer className="border-t bg-background/80 backdrop-blur-sm w-full py-3 px-4 md:px-12">
			<div className="flex flex-col space-y-3 md:flex-row md:justify-between md:items-center md:space-y-0">
				{/* Copyright - Line 1 on mobile */}
				<div className="text-sm text-muted-foreground text-center md:text-left">
					© {new Date().getFullYear()} MediUX. All rights reserved.
				</div>

				{/* Links row - Line 2 on mobile */}
				<div className="flex justify-center space-x-4 items-center">
					<Link
						href="https://mediux.io"
						target="_blank"
						rel="noopener noreferrer"
						className="group flex items-center hover:text-primary transition-colors active:scale-95 hover:brightness-120"
					>
						MediUX
						<div className="relative ml-1 w-[16px] h-[16px] rounded-t-md overflow-hidden">
							<Image src="/mediux_logo.svg" alt="Logo" fill className="object-contain" priority />
						</div>
					</Link>

					<Link
						href="https://discord.gg/YAKzwKPwyw"
						target="_blank"
						rel="noopener noreferrer"
						className="flex items-center hover:text-primary transition-colors active:scale-95 hover:brightness-120"
					>
						Discord <FaDiscord className="ml-1 h-3 w-3" />
					</Link>

					<Link
						href="https://github.com/mediux-team/aura"
						target="_blank"
						rel="noopener noreferrer"
						className="flex items-center hover:text-primary transition-colors active:scale-95 hover:brightness-120"
					>
						GitHub <FaGithub className="ml-1 h-3 w-3" />
					</Link>
				</div>

				{/* Version - Line 3 on mobile */}
				<div className="flex flex-col items-center md:flex-row md:justify-end gap-2 hover:brightness-120 active:scale-95 transition">
					<Link
						href={`/change-log?currentVersion=${encodeURIComponent(version)}`}
						className="text-sm py-1 px-2 bg-muted rounded-md hover:text-primary transition-colors"
						title="View change log"
					>
						App Version: {version}
					</Link>
					{latestVersion && isNewerVersion(latestVersion, version) && (
						<Link
							href={`/change-log?currentVersion=${encodeURIComponent(version)}&updates=true&latestVersion=${encodeURIComponent(latestVersion)}`}
							className="text-sm py-1 px-2 rounded-md bg-amber-500/10 border border-amber-500/40 hover:bg-amber-500/20 transition-colors"
							title={`Change log for latest version ${latestVersion} available`}
						>
							Update Available: {latestVersion}
						</Link>
					)}
				</div>
			</div>
			<ReleaseNotesDialog open={showReleaseNotes} onOpenChange={handleCloseReleaseNotes} changelog={changelog} />
		</footer>
	);
}

export default AppFooter;
