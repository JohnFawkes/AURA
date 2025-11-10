"use client";

import { GalleryHorizontalEnd } from "lucide-react";

import { createContext, useContext, useEffect, useState } from "react";

import { Slider } from "@/components/ui/slider";

import { cn } from "@/lib/cn";

export type DensityStep = 0 | 1 | 2;

interface ViewDensityContextType {
	densityStep: DensityStep; // 0 = highest density, 1 = medium density, 2 = lowest density
	setDensityStep: (value: DensityStep) => void;
}

const ViewDensityContext = createContext<ViewDensityContextType | undefined>(undefined);

interface ViewDensityProviderProps {
	children: React.ReactNode;
	defaultStep?: DensityStep;
}

export function ViewDensityProvider({ children, defaultStep = 1 }: ViewDensityProviderProps) {
	const [densityStep, setDensityStep] = useState<DensityStep>(defaultStep);

	// Persist density preference in localStorage
	useEffect(() => {
		const savedDensity = localStorage.getItem("viewDensityPreference");
		if (savedDensity) {
			const parsedStep = parseInt(savedDensity, 10) as DensityStep;
			// Ensure the value is within our allowed steps
			if (parsedStep >= 0 && parsedStep <= 2) {
				setDensityStep(parsedStep);
			}
		}
	}, []);

	useEffect(() => {
		localStorage.setItem("viewDensityPreference", densityStep.toString());
	}, [densityStep]);

	return (
		<ViewDensityContext.Provider value={{ densityStep, setDensityStep }}>{children}</ViewDensityContext.Provider>
	);
}

export function useViewDensity() {
	const context = useContext(ViewDensityContext);
	if (context === undefined) {
		throw new Error("useViewDensity must be used within a ViewDensityProvider");
	}
	return context;
}

export function ViewDensitySlider({ className }: { className?: string }) {
	const { densityStep, setDensityStep } = useViewDensity();

	return (
		<div className={cn("flex items-center gap-2", className)}>
			<GalleryHorizontalEnd className="h-4 w-4 text-muted-foreground" />
			<Slider
				className="w-46"
				value={[densityStep]}
				min={0}
				max={2}
				step={1}
				onValueChange={(values) => setDensityStep(values[0] as DensityStep)}
				aria-label="Adjust view density"
			/>
		</div>
	);
}
