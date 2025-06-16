export const log = (...args: unknown[]) => {
	const debugMode = localStorage.getItem("debugMode") === "true";
	if (debugMode) {
		// eslint-disable-next-line no-console
		console.log(...args);
	}
};
