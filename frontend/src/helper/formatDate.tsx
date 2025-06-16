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
		} else {
			return `${pluralize(diffDays, "day")} ago`;
		}
	} catch {
		return "Invalid Date";
	}
};
