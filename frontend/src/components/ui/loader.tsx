import React from "react";
import { LoaderIcon } from "lucide-react";

const Loader: React.FC<{ message: string }> = ({ message }) => {
	return (
		<div className="flex flex-col items-center justify-center">
			<LoaderIcon className="animate-spin h-6 w-6 text-gray-500" />
			{message && (
				<span className="mt-2 text-gray-500 text-center">
					{message}
				</span>
			)}
		</div>
	);
};

export default Loader;
