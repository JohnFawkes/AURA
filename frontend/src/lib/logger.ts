/* eslint-disable no-console */
/* eslint-disable @typescript-eslint/no-unused-expressions */
export function log(level: "INFO" | "WARN" | "ERROR", page: string, action: string, message: string, data?: unknown) {
	const debugMode =
		(typeof window !== "undefined" && localStorage.getItem("debugMode") === "true") ||
		process.env.NEXT_PUBLIC_APP_VERSION?.endsWith("dev");

	if (!debugMode && level !== "ERROR" && level !== "WARN") return;

	const now = new Date();
	const pad = (n: number, len = 2) => n.toString().padStart(len, "0");
	const formattedTime = `${now.getFullYear()}/${pad(now.getMonth() + 1)}/${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}:${pad(now.getMilliseconds(), 3)}`;

	const base = `[${formattedTime}] [${level}] [${page}] [${action}] - ${message}`;
	if (level === "ERROR") {
		data !== undefined ? console.error(base, data) : console.error(base);
	} else if (level === "WARN") {
		data !== undefined ? console.warn(base, data) : console.warn(base);
	} else if (level === "INFO") {
		data !== undefined ? console.info(base, data) : console.info(base);
	} else {
		data !== undefined ? console.log(base, data) : console.log(base);
	}
}
