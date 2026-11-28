// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useRef, useState } from "react";

import axios from "axios";
import { useTranslation } from "react-i18next";

import { Alert, AlertTitle } from "@components/UI/Alert";
import { Button } from "@components/UI/Button";
import { FloatingInput } from "@components/UI/FloatingInput";
import { PasswordVisibilityToggle } from "@components/UI/PasswordVisibilityToggle";
import { Spinner } from "@components/UI/Spinner";
import { useNotifications } from "@contexts/NotificationsContext";
import { usePasswordVisibility } from "@hooks/PasswordVisibility";
import LoginLayout from "@layouts/LoginLayout";
import { postPasswordChange } from "@services/ChangePassword";
import { signOut } from "@services/SignOut";

export interface Props {
    username: string;

    onPasswordChanged: () => void;
    onSignedOut: () => void;
}

const PasswordChangeRequiredForm = function (props: Props) {
    const { t: translate } = useTranslation();
    const { createErrorNotification } = useNotifications();

    const [oldPassword, setOldPassword] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [repeatPassword, setRepeatPassword] = useState("");
    const [loading, setLoading] = useState(false);
    const [changed, setChanged] = useState(false);

    const [oldError, setOldError] = useState(false);
    const [newError, setNewError] = useState(false);
    const [repeatError, setRepeatError] = useState(false);

    const oldRef = useRef<HTMLInputElement | null>(null);
    const newRef = useRef<HTMLInputElement | null>(null);
    const repeatRef = useRef<HTMLInputElement | null>(null);

    const { showPassword, toggleProps } = usePasswordVisibility();

    const handleSubmit = useCallback(async () => {
        if (loading) {
            return;
        }

        if (changed) {
            props.onPasswordChanged();

            return;
        }

        setOldError(false);
        setNewError(false);
        setRepeatError(false);

        if (oldPassword === "") {
            setOldError(true);
            oldRef.current?.focus();

            return;
        }

        if (newPassword === "") {
            setNewError(true);
            newRef.current?.focus();

            return;
        }

        if (newPassword !== repeatPassword) {
            setRepeatError(true);
            repeatRef.current?.focus();
            createErrorNotification(translate("Passwords do not match"));

            return;
        }

        setLoading(true);

        try {
            await postPasswordChange(props.username, oldPassword, newPassword);

            setChanged(true);
            setLoading(false);
            props.onPasswordChanged();
        } catch (err) {
            const status = axios.isAxiosError(err) ? err.response?.status : undefined;

            if (status === 401) {
                setOldError(true);
                oldRef.current?.focus();
                createErrorNotification(translate("Incorrect password"));
            } else if (status === 400) {
                setNewError(true);
                setRepeatError(true);
                newRef.current?.focus();
                createErrorNotification(
                    translate("Your supplied password does not meet the password policy requirements"),
                );
            } else {
                createErrorNotification(translate("There was an issue changing the password"));
            }

            setLoading(false);
        }
    }, [changed, createErrorNotification, loading, newPassword, oldPassword, props, repeatPassword, translate]);

    const handleSignOut = useCallback(async () => {
        setLoading(true);

        try {
            await signOut(undefined);
        } catch {
            createErrorNotification(translate("There was an issue signing out"));
        }

        setLoading(false);
        props.onSignedOut();
    }, [createErrorNotification, props, translate]);

    return (
        <LoginLayout id="password-change-required" title={translate("Change your password")}>
            <div className="grid grid-cols-1 gap-4">
                <Alert>
                    <AlertTitle>{translate("Your password must be changed before you can continue")}</AlertTitle>
                </Alert>
                <div className="relative w-full">
                    <FloatingInput
                        ref={oldRef}
                        id="old-password"
                        label={translate("Current password")}
                        type={showPassword ? "text" : "password"}
                        error={oldError}
                        className="pr-10"
                        autoComplete="current-password"
                        autoFocus
                        disabled={changed}
                        value={oldPassword}
                        onChange={(e) => setOldPassword(e.target.value)}
                        onKeyDown={(e) => (e.key === "Enter" ? handleSubmit() : undefined)}
                    />
                    <PasswordVisibilityToggle
                        label={translate("Toggle password visibility")}
                        showPassword={showPassword}
                        {...toggleProps}
                    />
                </div>
                <FloatingInput
                    ref={newRef}
                    id="new-password"
                    label={translate("New password")}
                    type={showPassword ? "text" : "password"}
                    error={newError}
                    autoComplete="new-password"
                    disabled={changed}
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    onKeyDown={(e) => (e.key === "Enter" ? handleSubmit() : undefined)}
                />
                <FloatingInput
                    ref={repeatRef}
                    id="repeat-new-password"
                    label={translate("Repeat new password")}
                    type={showPassword ? "text" : "password"}
                    error={repeatError}
                    autoComplete="new-password"
                    disabled={changed}
                    value={repeatPassword}
                    onChange={(e) => setRepeatPassword(e.target.value)}
                    onKeyDown={(e) => (e.key === "Enter" ? handleSubmit() : undefined)}
                />
                <Button id="password-change-button" disabled={loading} onClick={handleSubmit}>
                    {loading ? <Spinner /> : translate(changed ? "Retry" : "Change password")}
                </Button>
                <Button
                    id="password-change-sign-out-button"
                    variant="outline"
                    disabled={loading}
                    onClick={handleSignOut}
                >
                    {translate("Sign out")}
                </Button>
            </div>
        </LoginLayout>
    );
};

export default PasswordChangeRequiredForm;
