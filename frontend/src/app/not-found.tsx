import Link from "next/link";

export default function Custom404() {
	return (
		<div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-700 px-4">
			<div className="text-center">
				<h1 className="text-7xl md:text-9xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-pink-500 to-purple-400 mb-4">
					404
				</h1>
				<h2 className="text-2xl md:text-3xl font-semibold text-white mb-2">Page Not Found</h2>
				<p className="text-gray-300 mb-8">
					Sorry, the page you are looking for does not exist or has been moved.
				</p>
				<Link
					href="/"
					className="inline-block px-6 py-3 rounded bg-pink-600 text-white font-semibold shadow hover:bg-pink-700 transition"
				>
					Go Home
				</Link>
			</div>
		</div>
	);
}
