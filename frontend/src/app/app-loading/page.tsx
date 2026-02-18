"use client";
import React, { useEffect } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { H2 } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { useOnboardingStore } from "@/lib/stores/global-store-onboarding";

const boolBadge = (val: boolean | undefined) => (
  <Badge className={cn("ml-2", val === true ? "bg-green-500 text-white" : "bg-red-500 text-white")}>
    {val ? "Yes" : "No"}
  </Badge>
);

const AppStatusPage: React.FC = () => {
  // Get Onboarding Status
  const status = useOnboardingStore((state) => state.status);
  const hasHydrated = useOnboardingStore((state) => state.hasHydrated);
  const fetchStatus = useOnboardingStore((state) => state.fetchStatus);

  // Fetch status on mount
  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  useEffect(() => {
    // If we have hydrated and the app is fully loaded, redirect to dashboard
    if (
      hasHydrated &&
      status?.app_fully_loaded &&
      !status?.needs_setup &&
      status?.config_valid &&
      status?.config_loaded
    ) {
      window.location.href = "/";
    }
  }, [hasHydrated, status]);

  return (
    <div className="container mx-auto p-4 min-h-screen flex flex-col items-center">
      <H2 className="text-3xl font-bold mb-4">App Status</H2>

      {!hasHydrated ? (
        <div>Loading...</div>
      ) : (
        <div className="w-full max-w-4xl mb-4 p-3 rounded text-sm whitespace-pre-wrap border">
          <div className="mb-2">
            <b>Media Server Name:</b> {status?.media_server_name || "N/A"}
          </div>
          <div className="mb-2">
            <b>App Fully Loaded:</b> {boolBadge(status?.app_fully_loaded)}
          </div>
          <div className="mb-2">
            <b>Config Loaded:</b> {boolBadge(status?.config_loaded)}
          </div>
          <div className="mb-2">
            <b>Config Valid:</b> {boolBadge(status?.config_valid)}
          </div>
          <div className="mb-2">
            <b>Needs Setup:</b> {boolBadge(status?.needs_setup)}
          </div>

          <Button
            variant="outline"
            onClick={() => {
              fetchStatus();
            }}
          >
            Refresh Status
          </Button>
        </div>
      )}
    </div>
  );
};

export default AppStatusPage;
