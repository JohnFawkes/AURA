"use client";

import { LoaderIcon } from "lucide-react";
import { toast } from "sonner";

import React, { useCallback, useEffect, useRef, useState } from "react";

import { useRouter } from "next/navigation";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { H2 } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { useOnboardingStore } from "@/lib/stores/global-store-onboarding";

const boolBadge = (val: boolean | undefined, invert = false) => {
  const normalized = val === true;
  const isGood = invert ? !normalized : normalized;

  return (
    <Badge className={cn("ml-2", isGood ? "bg-green-500 text-white" : "bg-red-500 text-white")}>
      {normalized ? "Yes" : "No"}
    </Badge>
  );
};

const POLL_INTERVAL_SECONDS = 3;
const POLL_INTERVAL_MS = POLL_INTERVAL_SECONDS * 1000;

// Hold for 1s at the end before restarting
const RESTART_DELAY_MS = 1000;
const CYCLE_INTERVAL_MS = POLL_INTERVAL_MS + RESTART_DELAY_MS;

// Progress fill speed
const PROGRESS_PER_SECOND = 30;

const AppStatusPage: React.FC = () => {
  const router = useRouter();

  const status = useOnboardingStore((state) => state.status);
  const hasHydrated = useOnboardingStore((state) => state.hasHydrated);
  const fetchStatus = useOnboardingStore((state) => state.fetchStatus);

  const [nextRetryAt, setNextRetryAt] = useState(() => Date.now() + POLL_INTERVAL_MS);
  const [secondsLeft, setSecondsLeft] = useState(POLL_INTERVAL_SECONDS);
  const [progress, setProgress] = useState(0);
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const startPolling = useCallback(() => {
    if (pollIntervalRef.current) {
      clearInterval(pollIntervalRef.current);
    }

    pollIntervalRef.current = setInterval(() => {
      const next = Date.now() + CYCLE_INTERVAL_MS;
      setNextRetryAt(next);
      void fetchStatus();
    }, CYCLE_INTERVAL_MS);
  }, [fetchStatus]);

  const refreshNow = useCallback(() => {
    const next = Date.now() + CYCLE_INTERVAL_MS;
    setNextRetryAt(next);
    void fetchStatus();

    // Restart polling so next auto retry is a full cycle from this click.
    startPolling();
  }, [fetchStatus, startPolling]);

  useEffect(() => {
    refreshNow();
    return () => {
      if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
    };
  }, [refreshNow]);

  useEffect(() => {
    const timerId = setInterval(() => {
      const now = Date.now();
      const remainingMs = Math.max(0, nextRetryAt - now);

      // Countdown only covers the first 3 seconds; final 1 second is hold-at-end.
      const countdownMs = Math.max(0, remainingMs - RESTART_DELAY_MS);
      setSecondsLeft(Math.ceil(countdownMs / 1000));

      // Elapsed within the active fill window (0..POLL_INTERVAL_MS)
      const elapsedMs = Math.min(POLL_INTERVAL_MS, POLL_INTERVAL_MS - countdownMs);

      // 30% per second, clamped to 100
      const nextProgress = Math.min(100, (elapsedMs / 1000) * PROGRESS_PER_SECOND);

      // During hold second, pin to 100 so bar reaches end and stays there briefly.
      setProgress(countdownMs === 0 ? 100 : nextProgress);
    }, 100);

    return () => clearInterval(timerId);
  }, [nextRetryAt]);

  useEffect(() => {
    if (!hasHydrated || !status) return;

    const isReady = status.app_fully_loaded && !status.needs_setup && status.config_valid && status.config_loaded;
    if (isReady) {
      toast.dismiss("app-loading-toast");
      router.replace("/");
    }
  }, [hasHydrated, status, router]);

  useEffect(() => {
    if (!hasHydrated || !status) return;
    if (status.app_fully_loaded) return;
  }, [hasHydrated, status, secondsLeft]);

  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-muted/30 flex items-center justify-center p-4">
      <div className="w-full max-w-4xl rounded-2xl border bg-card/80 backdrop-blur p-6 shadow-xl">
        <div className="mb-6">
          <H2 className="text-3xl font-bold tracking-tight">Aura - App Status</H2>
          <p className="text-sm text-muted-foreground mt-1">
            App is still loading. Trying again in <span className="font-semibold">{secondsLeft}s</span>
            {progress > 0 && progress < 100 && (
              <>
                <LoaderIcon className="inline-block animate-spin ml-2 mb-2" size={16} />
              </>
            )}
          </p>
        </div>

        <div className="mb-6">
          <div className="h-2 rounded-full bg-muted overflow-hidden">
            <div
              className="h-full bg-primary transition-[width] duration-100 ease-linear"
              style={{ width: `${progress}%` }}
            />{" "}
          </div>
        </div>

        {!hasHydrated ? (
          <div className="text-sm text-muted-foreground">Loading status...</div>
        ) : (
          <div className="w-full rounded-lg text-sm border overflow-hidden">
            <table className="w-full border-collapse">
              <tbody>
                <tr className="border-b">
                  <td className="py-3 px-4 font-medium text-muted-foreground w-56">Media Server Name</td>
                  <td className="py-3 px-4 text-right">{status?.media_server_name || "N/A"}</td>
                </tr>
                <tr className="border-b">
                  <td className="py-3 px-4 font-medium text-muted-foreground">App Version</td>
                  <td className="py-3 px-4 text-right">{status?.app_version || "N/A"}</td>
                </tr>
                <tr className="border-b">
                  <td className="py-3 px-4 font-medium text-muted-foreground">App Fully Loaded</td>
                  <td className="py-3 px-4 text-right">{boolBadge(status?.app_fully_loaded)}</td>
                </tr>
                <tr className="border-b">
                  <td className="py-3 px-4 font-medium text-muted-foreground">Config Loaded</td>
                  <td className="py-3 px-4 text-right">{boolBadge(status?.config_loaded)}</td>
                </tr>
                <tr className="border-b">
                  <td className="py-3 px-4 font-medium text-muted-foreground">Config Valid</td>
                  <td className="py-3 px-4 text-right">{boolBadge(status?.config_valid)}</td>
                </tr>
                <tr>
                  <td className="py-3 px-4 font-medium text-muted-foreground">Needs Setup</td>
                  <td className="py-3 px-4 text-right">{boolBadge(status?.needs_setup, true)}</td>
                </tr>
                <tr>
                  <td className="py-3 px-4 font-medium text-muted-foreground">Current Step</td>
                  <td className="py-3 px-4 text-right">
                    <Badge
                      className={cn(
                        "ml-2",
                        status?.app_loading_step === "Initializing Application"
                          ? "bg-yellow-500 text-white"
                          : status?.app_loading_step === "Bootstrapping Application"
                            ? "bg-blue-500 text-white"
                            : status?.app_loading_step === "Performing Pre-Flight Checks"
                              ? "bg-purple-500 text-white"
                              : status?.app_loading_step === "Warming Up Application"
                                ? "bg-indigo-500 text-white"
                                : status?.app_loading_step === "App Fully Loaded"
                                  ? "bg-green-500 text-white"
                                  : "bg-gray-500 text-white"
                      )}
                    >
                      {status?.app_loading_step || "N/A"}
                    </Badge>
                  </td>
                </tr>
              </tbody>
            </table>

            <div className="p-4 border-t bg-muted/20 flex justify-end">
              <Button variant="outline" onClick={refreshNow}>
                Refresh Status
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default AppStatusPage;
