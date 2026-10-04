// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { KeyboardEvent, useCallback, useEffect, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { PasswordVisibilityToggle } from "@components/UI/PasswordVisibilityToggle";
import { Spinner } from "@components/UI/Spinner";
import { useNotifications } from "@contexts/NotificationsContext";
import { usePasswordVisibility } from "@hooks/PasswordVisibility";
import { postFirstFactorReauthenticate } from "@services/Password";

export interface Props {
    onAuthenticationSuccess: () => void;
}

const ReauthenticationPasswordForm = function (props: Props) {
    const { onAuthenticationSuccess } = props;
    const { t: translate } = useTranslation(["portal", "settings"]);
    const { createErrorNotification } = useNotifications();
    const { showPassword, toggleProps } = usePasswordVisibility();

    const [loading, setLoading] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordError, setPasswordError] = useState(false);

    const passwordRef = useRef<HTMLInputElement | null>(null);

    useEffect(() => {
        const timeout = setTimeout(() => passwordRef.current?.focus(), 10);
        return () => clearTimeout(timeout);
    }, []);

    const handleSubmit = useCallback(async () => {
        if (password === "") {
            setPasswordError(true);
            passwordRef.current?.focus();

            return;
        }

        setLoading(true);

        try {
            await postFirstFactorReauthenticate(password);
            onAuthenticationSuccess();
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Incorrect password"));
            setPassword("");
            setLoading(false);
            passwordRef.current?.focus();
        }
    }, [createErrorNotification, onAuthenticationSuccess, password, translate]);

    const handleKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            if (event.key !== "Enter") return;

            event.preventDefault();
            handleSubmit().catch(console.error);
        },
        [handleSubmit],
    );

    return (
        <form id={"form-reauthenticate-password"} onSubmit={(e) => e.preventDefault()}>
            <div className="grid grid-cols-1 gap-4">
                <div className="w-full">
                    <Label htmlFor="reauthenticate-password-textfield">{translate("Password")} *</Label>
                    <div className="relative">
                        <Input
                            ref={passwordRef}
                            id="reauthenticate-password-textfield"
                            required
                            disabled={loading}
                            value={password}
                            className="pr-10"
                            error={passwordError}
                            onChange={(v) => setPassword(v.target.value)}
                            onFocus={() => setPasswordError(false)}
                            type={showPassword ? "text" : "password"}
                            autoComplete="current-password"
                            onKeyDown={handleKeyDown}
                        />
                        <PasswordVisibilityToggle
                            label={translate("Toggle password visibility")}
                            showPassword={showPassword}
                            {...toggleProps}
                        />
                    </div>
                </div>
                <div className="w-full">
                    <Button
                        id="reauthenticate-button"
                        variant="default"
                        className="w-full"
                        disabled={loading}
                        onClick={() => {
                            handleSubmit().catch(console.error);
                        }}
                    >
                        {translate("Authenticate", { ns: "settings" })}
                        {loading ? <Spinner size={20} className="ml-2 h-5 w-5" /> : null}
                    </Button>
                </div>
            </div>
        </form>
    );
};

export default ReauthenticationPasswordForm;
