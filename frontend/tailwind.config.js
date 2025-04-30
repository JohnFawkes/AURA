module.exports = {
	darkMode: "class", // Use class-based dark mode
	content: ["./src/**/*.{js,ts,jsx,tsx}"], // Specify the files Tailwind should scan
	theme: {
		extend: {
			colors: {
				primary: {
					DEFAULT: "#1E3A8A", // Tailwind's blue-800
					light: "#3B82F6", // Tailwind's blue-500
					dark: "#1E40AF", // Tailwind's blue-900
				},
				secondary: {
					DEFAULT: "#FBBF24", // Tailwind's yellow-400
					light: "#FCD34D", // Tailwind's yellow-300
					dark: "#F59E0B", // Tailwind's yellow-500
				},
				background: {
					DEFAULT: "#F3F4F6", // Tailwind's gray-100
					light: "#FFFFFF", // Tailwind's white
					dark: "#111827", // Tailwind's gray-900
				},
				text: {
					DEFAULT: "#111827", // Tailwind's gray-900
					light: "#374151", // Tailwind's gray-700
					dark: "#F3F4F6", // Tailwind's gray-100
				},
			},
		},
	},
	plugins: [],
};
