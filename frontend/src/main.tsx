import { StrictMode, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App.tsx";
import { StyledEngineProvider, ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";

import { lightTheme, darkTheme } from "./theme";

import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";

import "slick-carousel/slick/slick.css";
import "slick-carousel/slick/slick-theme.css";

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<BrowserRouter>
			<ThemeWrapper />
		</BrowserRouter>
	</StrictMode>
);

function ThemeWrapper() {
	const [darkMode, setDarkMode] = useState(false);

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
}
