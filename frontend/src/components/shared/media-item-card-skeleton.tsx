"use client";

import Image from "next/image";

import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

import { cn } from "@/lib/cn";

// Single skeleton card
export const HomeMediaItemCardSkeleton: React.FC = () => (
  <Card className="relative items-center cursor-pointer hover:shadow-xl transition-shadow">
    <div className="absolute top-2 left-2 z-10">
      <Skeleton className="h-5 w-5 rounded-full bg-green-500" />
    </div>

    {/* Poster Image */}
    <Image
      src="/aura_logo.svg"
      alt="Aura Logo"
      width={136}
      height={204}
      className="opacity-70 animate-pulse w-[80%] h-auto"
      draggable={false}
    />

    {/* Title */}
    <span className="text-center text-lg text-foreground font-semibold mb-0">
      <Skeleton className="h-6 w-[80%] mx-auto mb-2" />
    </span>

    {/* Badges */}
    <CardContent className="flex flex-col md:flex-row justify-center items-center gap-1 p-1">
      <Skeleton className="h-5 w-12 rounded" />
      <Skeleton className="h-5 w-20 rounded" />
    </CardContent>
  </Card>
);

// Grid of skeleton cards, centered and responsive
export const HomeMediaItemCardSkeletonGrid: React.FC = () => {
  const cols = "grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5";
  return (
    <div className="flex justify-center">
      <div className={cn("grid gap-4", cols)}>
        {Array.from({ length: 10 }).map((_, i) => (
          <div key={i} className="flex flex-col gap-1.5">
            <HomeMediaItemCardSkeleton />
          </div>
        ))}
      </div>
    </div>
  );
};
