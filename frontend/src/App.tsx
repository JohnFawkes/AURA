import "./App.css";
import { Route, Routes } from "react-router-dom";
//import Home from "./pages/s";
import Home from "./pages/Home";
import PlexDetails from "./pages/PlexDetails";
import PageNotFound from "./pages/PageNotFound";

function App() {
	return (
		<div className="App">
			<Routes>
				<Route path="/" element={<Home />} />
				<Route path="/plex" element={<PlexDetails />} />

				<Route path="*" element={<PageNotFound />} />
			</Routes>
		</div>
	);
}

export default App;
