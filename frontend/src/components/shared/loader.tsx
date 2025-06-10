import { LoaderIcon } from "lucide-react";

import React from "react";

import { cn } from "@/lib/utils";

interface LoaderProps {
	message?: string;
	className?: string;
	iconClassName?: string;
	messageClassName?: string;
}

const Loader: React.FC<LoaderProps> = ({ message, className, iconClassName, messageClassName }) => {
	return (
		<div className={cn("flex flex-col items-center justify-center", className)}>
			<LoaderIcon className={cn("animate-spin h-6 w-6 text-gray-500", iconClassName)} />
			{message && (
				<span className={cn("mt-2 text-gray-500 text-center", messageClassName)}>
					{message}
				</span>
			)}
		</div>
	);
};

export default Loader;
