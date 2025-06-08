import localFont from "next/font/local";

// Fonts included in Gabarito folder
// - Black
// - Bold
// - ExtraBold
// - Medium
// - Regular
// - SemiBold
// - Variable

export const gabarito = localFont({
	src: [
		{
			path: "./Gabarito/Gabarito-Black.ttf",
			weight: "900",
			style: "normal",
		},
		{
			path: "./Gabarito/Gabarito-Bold.ttf",
			weight: "700",
			style: "normal",
		},
		{
			path: "./Gabarito/Gabarito-ExtraBold.ttf",
			weight: "800",
			style: "normal",
		},
		{
			path: "./Gabarito/Gabarito-Medium.ttf",
			weight: "500",
			style: "normal",
		},
		{
			path: "./Gabarito/Gabarito-Regular.ttf",
			weight: "400",
			style: "normal",
		},
		{
			path: "./Gabarito/Gabarito-SemiBold.ttf",
			weight: "600",
			style: "normal",
		},
		{
			path: "./Gabarito/Gabarito-VariableFont_wght.ttf",
			weight: "100 900",
			style: "normal",
		},
	],
	display: "swap",
	variable: "--font-gabarito",
});
