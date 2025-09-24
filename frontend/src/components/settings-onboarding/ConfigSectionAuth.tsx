import React, { useEffect, useRef } from "react";

import Link from "next/link";

import { PopoverHelp } from "@/components/shared/popover-help";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

import { cn } from "@/lib/cn";

import { AppConfigAuth } from "@/types/config/config-app";

interface ConfigSectionAuthProps {
	value: AppConfigAuth;
	editing: boolean;
	dirtyFields?: { Enabled?: boolean; Password?: boolean };
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
	const password = value.Password.trim();
	const passwordRequired = value.Enabled;
	const hasPassword = password.length > 0;
	const hashFormatValid = !hasPassword || hashRegex.test(password);
	const missingRequired = passwordRequired && !hasPassword;
	const valid = !missingRequired && hashFormatValid;
	const prevErrorRef = useRef<string | null>(null);

	let errorMsg: string | null = null;
	if (missingRequired) {
		errorMsg = "Password hash required when authentication is enabled.";
	} else if (hasPassword && !hashFormatValid) {
		errorMsg = "Invalid argon2id hash format.";
	}

	useEffect(() => {
		if (!errorsUpdate) return;

		// Only notify if error string actually changed
		if (prevErrorRef.current === errorMsg) return;
		prevErrorRef.current = errorMsg;

		if (errorMsg) {
			errorsUpdate({ Password: errorMsg });
		} else {
			// send empty to clear
			errorsUpdate({});
		}
	}, [errorMsg, errorsUpdate]);

	return (
		<Card className="p-5 space-y-1">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Authentication</h2>
			</div>

			<div
				className={cn(
					"flex items-center justify-between border rounded-md p-3 transition",
					"border-muted",
					dirtyFields.Enabled && "border-amber-500"
				)}
			>
				<Label className="mr-2">Enabled</Label>
				<div className="flex items-center gap-2">
					<Switch
						disabled={!editing}
						checked={value.Enabled}
						onCheckedChange={(c) => onChange("Enabled", c)}
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

			<div className="flex items-start gap-2">
				<div
					className={cn(
						"relative flex-1 border rounded-md p-3 space-y-2 transition",
						(dirtyFields.Password || errorMsg) && "ring-0",
						errorMsg ? "border-red-500" : dirtyFields.Password ? "border-amber-500" : "border-muted"
					)}
				>
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
						value={value.Password}
						onChange={(e) => onChange("Password", e.target.value)}
						aria-invalid={!valid}
					/>
					{errorMsg && <p className="text-xs text-red-500">{errorMsg}</p>}
				</div>
			</div>
		</Card>
	);
};
