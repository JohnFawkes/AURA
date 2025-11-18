export const formatLastUpdatedDate = (lastUpdateString: string, dateCreatedString: string) => {
	try {
		if (!lastUpdateString && !dateCreatedString) {
			return "Invalid Date";
		}
		let useDateString = lastUpdateString;
		if (lastUpdateString === "0001-01-01T00:00:00Z") {
			useDateString = dateCreatedString;
		}
		const date = new Date(useDateString);
		if (isNaN(date.getTime())) throw new Error();

		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
		const diffWeeks = Math.floor(diffDays / 7);
		const diffMonths = Math.floor(diffDays / 30);
		const diffYears = Math.floor(diffDays / 365);

		const pluralize = (value: number, unit: string) => `${value} ${unit}${value !== 1 ? "s" : ""}`;

		if (diffYears >= 1) {
			return `Over ${pluralize(diffYears, "year")} ago`;
		} else if (diffMonths >= 1) {
			return `${pluralize(diffMonths, "month")} ago`;
		} else if (diffWeeks >= 1) {
			return `${pluralize(diffWeeks, "week")} ago`;
		} else if (diffDays >= 1) {
			return `${pluralize(diffDays, "day")} ago`;
		} else {
			const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
			const diffMinutes = Math.floor(diffMs / (1000 * 60));
			if (diffHours >= 1) {
				return `${pluralize(diffHours, "hour")} ago`;
			} else if (diffMinutes < 5) {
				return "Just a moment ago";
			} else {
				return `${pluralize(diffMinutes, "minute")} ago`;
			}
		}
	} catch {
		return "Invalid Date";
	}
};

export const formatExactDateTime = (dateString: string) => {
	try {
		if (!dateString) {
			return "Invalid Date";
		}
		const date = new Date(dateString);
		if (isNaN(date.getTime())) throw new Error();

		const pad = (n: number) => n.toString().padStart(2, "0");
		return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
	} catch {
		return "Invalid Date";
	}
};
