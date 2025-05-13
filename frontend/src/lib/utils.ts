import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

/**
 * Sets a CSS variable value on the document's root element
 * @param variableName CSS variable name (without --)
 * @param value Value to set
 */
export function setCssVariable(variableName: string, value: string): void {
	if (typeof document !== "undefined") {
		const name = variableName.startsWith("--")
			? variableName
			: `--${variableName}`;
		document.documentElement.style.setProperty(name, value);
	}
}

/**
 * Formats a date string into a cache buster string (YYYYMMDDhhmmss)
 * @param dateString The date string to format (e.g. "2023-12-24T07:16:05.000Z")
 * @returns Formatted string (e.g. "20231224071605")
 */
export function formatDateToCacheBuster(dateString: string): string {
	const date = new Date(dateString);
	return date
		.toISOString()
		.replace(/[-T:.Z]/g, "")
		.slice(0, 14); // YYYYMMDDhhmmss
}

/**
 * Formats a date into a human-readable format (e.g. "18th January 2024")
 * @param dateString The date string to format
 * @returns Formatted date string
 */
export function formatDateToReadable(dateString: string): string {
	const date = new Date(dateString);
	const day = date.getDate();
	const month = date.toLocaleString("default", { month: "long" });
	const year = date.getFullYear();

	// Add ordinal suffix to day
	const suffix =
		["th", "st", "nd", "rd"][
			day % 100 > 10 && day % 100 < 14 ? 0 : day % 10
		] || "th";

	return `${day}${suffix} ${month} ${year}`;
}
