import React, { useMemo, useState, useEffect } from "react";
import { StyledEngineProvider, ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import App from "./App";
import { lightTheme, darkTheme } from "./theme";

const ThemeWrapper: React.FC = () => {
	// Initialize darkMode based on the user's device settings
	const [darkMode, setDarkMode] = useState(() => {
		return window.matchMedia("(prefers-color-scheme: dark)").matches;
	});

	// Listen for changes in the user's preferred color scheme
	useEffect(() => {
		const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
		const handleChange = (event: MediaQueryListEvent) => {
			setDarkMode(event.matches);
		};

		mediaQuery.addEventListener("change", handleChange);
		return () => mediaQuery.removeEventListener("change", handleChange);
	}, []);

	const theme = useMemo(
		() => (darkMode ? darkTheme : lightTheme),
		[darkMode]
	);

	return (
		<StyledEngineProvider injectFirst>
			<ThemeProvider theme={theme}>
				<CssBaseline />
				<App darkMode={darkMode} setDarkMode={setDarkMode} />
			</ThemeProvider>
		</StyledEngineProvider>
	);
};

export default ThemeWrapper;
