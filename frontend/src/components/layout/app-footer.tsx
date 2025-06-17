"use client";

import { FaDiscord, FaGithub } from "react-icons/fa";

import Image from "next/image";
import Link from "next/link";

interface AppFooterProps {
	version?: string;
}

export function AppFooter({ version = "dev" }: AppFooterProps) {
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
						href="https://mediux.pro"
						target="_blank"
						rel="noopener noreferrer"
						className="group flex items-center hover:text-primary transition-colors"
					>
						MediUX
						<div className="relative ml-1 w-[16px] h-[16px] rounded-t-md overflow-hidden">
							<Image src="/mediux_logo.svg" alt="Logo" fill className="object-contain" />
						</div>
					</Link>

					<Link
						href="https://discord.gg/HP9TpTmfcp"
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
				<div className="flex justify-center md:justify-end">
					<Link
						href="/logs"
						className="text-xs py-1 px-2 bg-muted rounded-md hover:text-primary transition-colors"
					>
						App Version: {version}
					</Link>
				</div>
			</div>
		</footer>
	);
}

export default AppFooter;
