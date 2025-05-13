"use client";
import React from "react";
import Link from "next/link";
import { FaGithub } from "react-icons/fa";
import { FaDiscord } from "react-icons/fa";
import Image from "next/image";

interface AppFooterProps {
	version?: string;
}

export function AppFooter({ version = "dev" }: AppFooterProps) {
	return (
		<footer className="border-t bg-background/80 backdrop-blur-sm w-full py-4 px-4 md:px-12">
			<div className="flex flex-col md:flex-row justify-between items-center space-y-4 md:space-y-0">
				<div className="text-sm text-muted-foreground">
					© {new Date().getFullYear()} MediUX. All rights reserved.
				</div>
				<Link
					href="https://mediux.pro"
					target="_blank"
					rel="noopener noreferrer"
					className="group flex items-center hover:text-primary transition-colors"
				>
					MediUX
					<div className="relative ml-1 w-[16px] h-[16px] rounded-t-md overflow-hidden">
						<Image
							src="/mediux.svg"
							alt="Logo"
							fill
							className="object-contain"
						/>
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
					href="https://github.com/xmoosex/poster-setter"
					target="_blank"
					rel="noopener noreferrer"
					className="flex items-center hover:text-primary transition-colors"
				>
					Github <FaGithub className="ml-1 h-3 w-3" />
				</Link>
				<Link
					href="/logs"
					className="text-xs py-1 px-2 bg-muted rounded-md hover:text-primary transition-colors"
				>
					App Version: {version}
				</Link>
			</div>
		</footer>
	);
}

export default AppFooter;
