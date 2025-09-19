// Regex for domain names: must have at least one dot, labels with letters, digits, hyphens, TLD at least 2 letters
const domainHostRegex = /^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$/;
// Regex for single-label hostnames (docker/container): letters, digits, hyphens only
const singleLabelHostRegex = /^[a-zA-Z0-9-]+$/;

/**
 * Validates a port number.
 * @param port Port as string or null
 * @param required Is port required for this connection type?
 * @param connectionType "ipv4", "docker", "domain"
 * @returns Error message or null if valid
 */
function validatePort(port: string | null, required: boolean, connectionType: string): string | null {
	if (!port) {
		// Allow default ports for http/https if not required
		if (!required) return null;
		switch (connectionType) {
			case "ipv4":
				return "Port is required for IPv4 addresses.";
			case "docker":
				return "Port is required for docker/container hostnames.";
			case "domain":
				// Domain names may omit port (default to 80/443)
				return null;
			default:
				return "Port is required.";
		}
	}
	const portNum = Number(port);
	if (!(portNum > 0 && portNum <= 65535)) {
		return `Port "${port}" is not valid. Must be between 1 and 65535.`;
	}
	return null;
}

/**
 * Validates an IPv4 address and its port.
 * @param host IPv4 address string
 * @param port Port string or null
 * @returns Error message or null if valid
 */
function validateIPv4Host(host: string, port: string | null): string | null {
	const errorMsg = `"${host}" is not a valid IPv4 address. Format: x.x.x.x, each between 0-255.`;
	if (!/^[0-9.]+$/.test(host)) return errorMsg;
	const parts = host.split(".");
	if (parts.length !== 4) return errorMsg;
	for (const p of parts) {
		if (p.length === 0 || (p.length > 1 && p.startsWith("0"))) return errorMsg;
		const n = Number(p);
		if (!Number.isInteger(n) || n < 0 || n > 255) return errorMsg;
	}
	return validatePort(port, true, "ipv4");
}

/**
 * Validates a domain name and its port.
 * @param host Domain name string
 * @param port Port string or null
 * @returns Error message or null if valid
 */
function validateDomainHost(host: string, port: string | null): string | null {
	if (!domainHostRegex.test(host)) {
		return `"${host}" is not a valid domain name. Example: example.com`;
	}
	return validatePort(port, false, "domain");
}

/**
 * Validates a docker/container hostname and its port.
 * @param host Hostname string
 * @param port Port string or null
 * @returns Error message or null if valid
 */
function validateDockerHost(host: string, port: string | null): string | null {
	if (!singleLabelHostRegex.test(host)) {
		return `"${host}" is not a valid docker/container host name. Only letters, numbers, and dashes allowed.`;
	}
	return validatePort(port, true, "docker");
}

/**
 * Validates a full media server URL.
 * Accepts http/https URLs for domain, IPv4, or docker hostnames.
 * @param raw Raw URL string
 * @returns Error message or null if valid
 */
export function ValidateURL(raw: string): string | null {
	const value = raw.trim();
	if (!/^https?:\/\//i.test(value)) {
		return "Must start with http:// or https://";
	}

	let parsed;
	try {
		parsed = new URL(value);
	} catch {
		return "Invalid URL format. Valid options are http://example.com, http://192.168.1.10:8080, http://my-docker-host:8080";
	}

	const protocol = parsed.protocol.toLowerCase();
	if (protocol !== "http:" && protocol !== "https:") {
		return "Only http and https protocols are allowed.";
	}

	const host = parsed.hostname;
	const port = parsed.port || null;

	// Case 1: Host looks like IPv4
	if (/^[0-9.]+$/.test(host)) {
		// Extract IP and port from raw input
		const ipMatch = value.match(/^https?:\/\/([0-9.]+)(?::\d+)?$/i);
		const portMatch = value.match(/^https?:\/\/[0-9.]+:(\d+)$/);
		const rawIp = ipMatch ? ipMatch[1] : host;
		const rawPort = portMatch ? portMatch[1] : port;
		return validateIPv4Host(rawIp, rawPort);
	}

	// Case 2: Host contains a dot (domain)
	if (host.includes(".")) {
		return validateDomainHost(host, port);
	}

	// Case 3: Single-label host (docker/container)
	const portMatch = value.match(/^https?:\/\/[a-zA-Z0-9-]+:(\d+)$/);
	const rawPort = portMatch ? portMatch[1] : port;
	return validateDockerHost(host, rawPort);
}
