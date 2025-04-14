import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App.tsx";
import { StyledEngineProvider } from "@mui/material/styles";

import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";

import "slick-carousel/slick/slick.css";
import "slick-carousel/slick/slick-theme.css";

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<BrowserRouter>
			{/* TODO: Add themes here for light and dark mode */}
			<StyledEngineProvider injectFirst>
				<App />
			</StyledEngineProvider>
		</BrowserRouter>
	</StrictMode>
);
