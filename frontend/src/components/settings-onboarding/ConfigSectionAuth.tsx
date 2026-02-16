import React, { useEffect, useMemo, useRef } from "react";

import Link from "next/link";

import { PopoverHelp } from "@/components/shared/popover-help";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import { AppConfigAuth } from "@/types/config/config";

interface ConfigSectionAuthProps {
    value: AppConfigAuth;
    editing: boolean;
    dirtyFields?: { enabled?: boolean; password?: boolean };
    onChange: <K extends keyof AppConfigAuth>(field: K, value: AppConfigAuth[K]) => void;
    errorsUpdate?: (errors: Partial<Record<keyof AppConfigAuth, string>>) => void;
}

const hashRegex = /^\$argon2id\$v=\d+\$m=\d+,t=\d+,p=\d+\$[A-Za-z0-9+/=]+\$[A-Za-z0-9+/=]+$/;

export const ConfigSectionAuth: React.FC<ConfigSectionAuthProps> = ({
    value,
    editing,
    dirtyFields = {},
    onChange,
    errorsUpdate,
}) => {
    const prevErrorRef = useRef<string>("");

    // Validation
    const errors = useMemo<Partial<Record<keyof AppConfigAuth, string>>>(() => {
        const errs: Partial<Record<keyof AppConfigAuth, string>> = {};
        // Password Errors
        if (value.enabled) {
            const password = value.password.trim();
            if (password.length === 0) {
                errs.password = "Password hash is required when authentication is enabled.";
            } else if (!hashRegex.test(password)) {
                errs.password = "Invalid Argon2id hash format.";
            }
        }
        return errs;
    }, [value.enabled, value.password]);

    // Emit errors upward
    useEffect(() => {
        if (!errorsUpdate) return;
        const serialized = JSON.stringify(errors);
        if (serialized === prevErrorRef.current) return;
        prevErrorRef.current = serialized;
        errorsUpdate(errors);
    }, [errors, errorsUpdate]);

    return (
        <Card className={`p-5 ${Object.values(errors).some(Boolean) ? "border-red-500" : "border-muted"}`}>
            <div className="flex items-center justify-between">
                <h2 className="text-xl font-semibold text-blue-500">Authentication</h2>
            </div>

            <div
                className={cn(
                    "flex items-center justify-between border rounded-md p-3 transition",
                    "border-muted",
                    dirtyFields.enabled && "border-amber-500"
                )}
            >
                <Label className="mr-2">Enabled</Label>
                <div className="flex items-center gap-2">
                    <Switch
                        disabled={!editing}
                        checked={value.enabled}
                        onCheckedChange={(c) => onChange("enabled", c)}
                    />
                    {editing && (
                        <PopoverHelp ariaLabel="help-auth-enabled">
                            <p>
                                Turn on to enforce authentication. A valid Argon2id password hash must be provided
                                below.
                            </p>
                        </PopoverHelp>
                    )}
                </div>
            </div>

            <div className="flex">
                <div className={cn("relative flex-1 border rounded-md p-3 space-y-2 transition")}>
                    <div>
                        <div className="flex items-center justify-between">
                            <Label htmlFor="auth-hash">Argon2id Password Hash</Label>
                            {editing && (
                                <PopoverHelp ariaLabel="help-auth-password-hash">
                                    <p className="mb-2">
                                        Provide an Argon2id hash. If authentication is enabled this hash must match the
                                        user's password.
                                    </p>
                                    <p>
                                        You can use a site like{" "}
                                        <Link
                                            className="text-primary underline"
                                            href="https://argon2.online/"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                        >
                                            Argon2.Online
                                        </Link>{" "}
                                        to generate a hash.
                                    </p>
                                </PopoverHelp>
                            )}
                        </div>
                        <Input
                            id="auth-hash"
                            disabled={!editing}
                            placeholder="$argon2id$v=19$m=65536,t=3,p=1$..."
                            type="text"
                            value={value.password}
                            onChange={(e) => onChange("password", e.target.value)}
                            className={cn("w-full mt-1", dirtyFields.password && "ring-2 ring-amber-500")}
                        />
                    </div>
                    {errors.password && <p className="text-xs text-red-500">{errors.password}</p>}
                </div>
            </div>
        </Card>
    );
};
