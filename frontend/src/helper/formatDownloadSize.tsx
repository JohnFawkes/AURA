export const formatDownloadSize = (downloadSizeBytes: number) => {
	if (downloadSizeBytes === 0) {
		return "";
	}
	const units = ["B", "KB", "MB", "GB", "TB"];
	const factor = 1024;
	const index = Math.floor(Math.log(downloadSizeBytes) / Math.log(factor));
	return `${(downloadSizeBytes / Math.pow(factor, index)).toFixed(2)} ${
		units[index]
	}`;
};
