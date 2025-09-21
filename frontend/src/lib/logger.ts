export const log = (...args: unknown[]) => {
	const debugMode =
		localStorage.getItem("debugMode") === "true" || process.env.NEXT_PUBLIC_APP_VERSION?.endsWith("dev");
	if (debugMode) {
		// eslint-disable-next-line no-console
		console.log(...args);
	}
};
