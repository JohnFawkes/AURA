"use client";

import cronstrue from "cronstrue";

import React, { useEffect, useRef } from "react";

import Link from "next/link";

import { PopoverHelp } from "@/components/shared/popover-help";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import { AppConfigAutoDownload } from "@/types/config/config";

interface ConfigSectionAutoDownloadProps {
  value: AppConfigAutoDownload;
  editing: boolean;
  dirtyFields?: Partial<Record<keyof AppConfigAutoDownload, boolean>>;
  onChange: <K extends keyof AppConfigAutoDownload>(field: K, value: AppConfigAutoDownload[K]) => void;
  errorsUpdate?: (errors: Partial<Record<keyof AppConfigAutoDownload, string>>) => void;
}

const validateCron = (expr: string): string | null => {
  if (!expr) return null;
  const trimmed = expr.trim();
  if (!trimmed) return "Cron expression is required when enabled.";
  try {
    cronstrue.toString(trimmed);
    return null;
  } catch {
    return "Invalid cron expression. Use a site like crontab.guru to help you create and test your cron expressions.";
  }
};

export const ConfigSectionAutoDownload: React.FC<ConfigSectionAutoDownloadProps> = ({
  value,
  editing,
  dirtyFields = {},
  onChange,
  errorsUpdate,
}) => {
  const prevErrorsRef = useRef<string>("{}");
  const errors = React.useMemo(() => {
    const errs: Partial<Record<keyof AppConfigAutoDownload, string>> = {};
    if (value.enabled) {
      const cronErr = validateCron(value.cron);
      if (cronErr) errs.cron = cronErr;
    }
    return errs;
  }, [value.enabled, value.cron]);

  useEffect(() => {
    if (!errorsUpdate) return;
    const ser = JSON.stringify(errors);
    if (ser === prevErrorsRef.current) return;
    prevErrorsRef.current = ser;
    errorsUpdate(errors);
  }, [errors, errorsUpdate]);

  return (
    <Card className={`p-5 ${Object.values(errors).some(Boolean) ? "border-red-500" : "border-muted"}`}>
      <h2 className="text-xl font-semibold text-blue-500">Auto Download</h2>

      {/* Enabled */}
      <div
        className={cn(
          "flex items-center justify-between border rounded-md p-3 transition",
          "border-muted",
          dirtyFields.enabled && "border-amber-500"
        )}
      >
        <Label className="mr-2">Enabled</Label>
        <div className="flex items-center gap-2">
          <Switch disabled={!editing} checked={value.enabled} onCheckedChange={(v) => onChange("enabled", v)} />
          {editing && (
            <PopoverHelp ariaLabel="help-auto-download-enabled">
              <p>
                Auto Download will check periodically for new updates to your saved sets. This is helpful if you want to
                download and apply season poster/titlecard updates from future updates to your sets.
              </p>
            </PopoverHelp>
          )}
        </div>
      </div>

      {/* Cron */}
      <div className={cn("space-y-1", "rounded-md")}>
        <div className="flex items-center justify-between">
          <Label>Cron Expression</Label>
          {editing && (
            <PopoverHelp ariaLabel="help-cron-expression">
              <p>
                Cron expression format. Use the standard cron syntax to specify when the job should run. You can use a
                site like{" "}
                <Link
                  className="text-primary hover:underline"
                  href="https://crontab.guru/"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  crontab.guru
                </Link>{" "}
                to help you create and test your cron expressions.
              </p>
            </PopoverHelp>
          )}
        </div>
        <Input
          disabled={!editing || !value.enabled}
          placeholder="e.g. 0 3 * * *"
          value={value.cron}
          onChange={(e) => onChange("cron", e.target.value)}
          className={cn(dirtyFields.cron && "border-amber-500")}
        />
        {errors.cron && <p className="text-xs text-red-500">{errors.cron}</p>}
        {value.enabled && !errors.cron && value.cron && value.cron.trim() && (
          <p className="text-xs text-muted-foreground">{cronstrue.toString(value.cron)}</p>
        )}
      </div>
    </Card>
  );
};
