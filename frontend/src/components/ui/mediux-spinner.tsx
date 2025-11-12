"use client";

import { cn } from "@/lib/cn";

interface AuraSpinnerProps {
	size?: "sm" | "md" | "lg" | "xl";
	className?: string;
}

export function AuraSpinner({ size = "md", className }: AuraSpinnerProps) {
	const sizeMap = {
		sm: 16,
		md: 24,
		lg: 32,
		xl: 48,
	};

	const logoSize = sizeMap[size];

	return (
		<div className={cn("relative", className)} style={{ width: logoSize, height: logoSize }}>
			<img src="/aura_logo.svg" width={logoSize} height={logoSize} alt="Loading..." className="animate-spin" />
		</div>
	);
}

export default AuraSpinner;
