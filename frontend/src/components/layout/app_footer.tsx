"use client";

import { useEffect, useState } from "react";
import { FaDiscord, FaGithub } from "react-icons/fa";

import Image from "next/image";
import Link from "next/link";

import { log } from "@/lib/logger";

interface AppFooterProps {
	version?: string;
}

export function AppFooter({ version = "dev" }: AppFooterProps) {
	// Get the latest version from Github
	const [latestVersion, setLatestVersion] = useState<string | null>(null);

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
				log("Error fetching latest version:", error);
			}
		};
		fetchLatestVersion();
	}, []);

	return (
		<footer className="border-t bg-background/80 backdrop-blur-sm w-full py-4 px-4 md:px-12">
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
						className="group flex items-center hover:text-primary transition-colors"
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
						className="flex items-center hover:text-primary transition-colors"
					>
						Discord <FaDiscord className="ml-1 h-3 w-3" />
					</Link>

					<Link
						href="https://github.com/mediux-team/aura"
						target="_blank"
						rel="noopener noreferrer"
						className="flex items-center hover:text-primary transition-colors"
					>
						GitHub <FaGithub className="ml-1 h-3 w-3" />
					</Link>
				</div>

				{/* Version - Line 3 on mobile */}
				<div className="flex flex-col items-center md:flex-row md:justify-end gap-2">
					<Link
						href="/logs"
						className="text-xs py-1 px-2 bg-muted rounded-md hover:text-primary transition-colors"
						title="View application logs"
					>
						App Version: {version}
					</Link>
					{latestVersion && latestVersion !== version && (
						<Link
							href="https://github.com/mediux-team/AURA/pkgs/container/aura"
							target="_blank"
							rel="noopener noreferrer"
							className="text-xs py-1 px-2 bg-amber-500/10 border border-amber-500/40 rounded-md hover:bg-amber-500/20 transition-colors"
							title={`Latest version ${latestVersion} on GitHub`}
						>
							Update Available: {latestVersion}
						</Link>
					)}
				</div>
			</div>
		</footer>
	);
}

export default AppFooter;
