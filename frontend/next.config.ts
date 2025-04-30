import type { NextConfig } from "next";

const APP_VERSION = process.env.NEXT_PUBLIC_APP_VERSION || "dev"; // Default to "dev" if not set

const nextConfig: NextConfig = {
	trailingSlash: true,
	images: {
		remotePatterns: [
			{
				protocol: "http",
				hostname: "10.1.1.30",
				port: "8888",
				pathname: "/**",
			},
			{
				protocol: "http",
				hostname: "localhost",
				port: "8888",
				pathname: "/**",
			},
		],
	},
	allowedDevOrigins: ["10.1.1.30", "localhost"],
	reactStrictMode: true,
	env: {
		NEXT_PUBLIC_APP_VERSION: APP_VERSION,
	},

	async rewrites() {
		return [
			{
				source: "/api/:path*",
				destination: `http://localhost:8888/api/:path*`,
			},
		];
	},
};

export default nextConfig;
