// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { KeyboardEvent, useActionState, useEffect, useEffectEvent, useRef, useState } from "react";

import { browserSupportsWebAuthn } from "@simplewebauthn/browser";
import { BroadcastChannel } from "broadcast-channel";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";

import { Alert, AlertTitle } from "@components/UI/Alert";
import { Button } from "@components/UI/Button";
import { Checkbox } from "@components/UI/Checkbox";
import { FloatingInput } from "@components/UI/FloatingInput";
import { Label } from "@components/UI/Label";
import { PasswordVisibilityToggle } from "@components/UI/PasswordVisibilityToggle";
import { Spinner } from "@components/UI/Spinner";
import { ResetPasswordStep1Route } from "@constants/Routes";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { usePasswordVisibility } from "@hooks/PasswordVisibility";
import { useQueryParam } from "@hooks/QueryParam";
import LoginLayout from "@layouts/LoginLayout";
import { IsCapsLockModified } from "@services/CapsLock";
import { postFirstFactor } from "@services/Password";
import PasskeyForm from "@views/LoginPortal/FirstFactor/PasskeyForm";

export interface Props {
    passkeyLogin: boolean;
    rememberMe: boolean;
    resetPassword: boolean;
    resetPasswordCustomURL: string;

    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
    onChannelStateChange: () => void;
}

const FirstFactorForm = function (props: Props) {
    const { t: translate } = useTranslation();

    const navigate = useNavigate();
    const redirectionURL = useQueryParam(RedirectionURL);
    const requestMethod = useQueryParam(RequestMethod);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();
    const { createErrorNotification } = useNotifications();
    const { showPassword, toggleProps } = usePasswordVisibility();

    const passkeyLogin = props.passkeyLogin && browserSupportsWebAuthn();

    const [rememberMe, setRememberMe] = useState(false);
    const [username, setUsername] = useState("");
    const [usernameError, setUsernameError] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordCapsLock, setPasswordCapsLock] = useState(false);
    const [passwordCapsLockPartial, setPasswordCapsLockPartial] = useState(false);
    const [passwordError, setPasswordError] = useState(false);
    const [passkeyAuthenticating, setPasskeyAuthenticating] = useState(false);

    const loginChannelRef = useRef<BroadcastChannel<boolean> | null>(null);
    const refocusPasswordRef = useRef(false);
    const usernameRef = useRef<HTMLInputElement | null>(null);
    const passwordRef = useRef<HTMLInputElement | null>(null);

    const focusUsername = () => {
        usernameRef.current?.focus();
    };

    const focusPassword = () => {
        passwordRef.current?.focus();
    };

    const handleChannelMessage = useEffectEvent((authenticated: boolean) => {
        if (authenticated) {
            props.onChannelStateChange();
        }
    });

    useEffect(() => {
        const channel = new BroadcastChannel<boolean>("login");

        loginChannelRef.current = channel;

        const handler = (authenticated: boolean) => handleChannelMessage(authenticated);

        channel.addEventListener("message", handler);

        return () => {
            channel.removeEventListener("message", handler);

            void channel.close();

            loginChannelRef.current = null;
        };
    }, []);

    const [, handleSignIn, isPending] = useActionState<null>(async () => {
        if (username === "" || password === "") {
            if (username === "") {
                setUsernameError(true);
            }

            if (password === "") {
                setPasswordError(true);
            }

            return null;
        }

        try {
            const res = await postFirstFactor(
                username,
                password,
                rememberMe,
                redirectionURL,
                requestMethod,
                flowID,
                flow,
                subflow,
                userCode,
            );

            await loginChannelRef.current?.postMessage(true);

            props.onAuthenticationSuccess(res ? res.redirect : undefined);
        } catch (err) {
            console.error(err);

            createErrorNotification(translate("Incorrect username or password"));

            setPassword("");

            // The inputs are disabled for the duration of the action, so the password field cannot take focus until
            // the pending state clears.
            refocusPasswordRef.current = true;
        }

        return null;
    }, null);

    const disabled = isPending || passkeyAuthenticating;

    useEffect(() => {
        if (disabled || !refocusPasswordRef.current) return;

        refocusPasswordRef.current = false;

        passwordRef.current?.focus();
    }, [disabled]);

    const handleRememberMeChange = () => {
        setRememberMe((prev) => !prev);
    };

    const handleResetPasswordClick = () => {
        if (props.resetPassword) {
            if (props.resetPasswordCustomURL) {
                window.open(props.resetPasswordCustomURL);
            } else {
                navigate(ResetPasswordStep1Route);
            }
        }
    };

    const handlePasswordKeyUp = (event: KeyboardEvent<HTMLInputElement>) => {
        if (password.length <= 1) {
            setPasswordCapsLock(false);
            setPasswordCapsLockPartial(false);

            if (password.length === 0) {
                return;
            }
        }

        const modified = IsCapsLockModified(event);

        if (modified === null) return;

        if (modified) {
            setPasswordCapsLock(true);
        } else {
            setPasswordCapsLockPartial(true);
        }
    };

    const handleUsernameKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key !== "Enter") return;

        event.preventDefault();

        if (disabled) return;

        if (!username.length) {
            setUsernameError(true);
        } else if (username.length && password.length) {
            handleSignIn();
        } else {
            setUsernameError(false);
            focusPassword();
        }
    };

    const handlePasswordKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key !== "Enter") return;

        event.preventDefault();

        if (disabled) return;

        if (!username.length) {
            focusUsername();
        } else if (!password.length) {
            focusPassword();
        }

        handleSignIn();
    };

    const handleRememberMeKeyDown = (event: KeyboardEvent<HTMLElement>) => {
        if (event.key !== "Enter") return;

        if (disabled) return;

        if (!username.length) {
            focusUsername();
        } else if (!password.length) {
            focusPassword();
        }

        handleSignIn();
    };

    return (
        <LoginLayout id="first-factor-stage" title={translate("Sign in")}>
            <form id={"form-login"} action={handleSignIn} noValidate>
                <div className="grid grid-cols-1 gap-5">
                    <div className="w-full">
                        <FloatingInput
                            ref={usernameRef}
                            id="username-textfield"
                            name="username"
                            label={`${translate("Username")} *`}
                            required
                            autoFocus
                            value={username}
                            error={usernameError}
                            disabled={disabled}
                            onChange={(v) => setUsername(v.target.value)}
                            onFocus={() => setUsernameError(false)}
                            autoCapitalize="none"
                            autoComplete={passkeyLogin ? "username webauthn" : "username"}
                            onKeyDown={handleUsernameKeyDown}
                        />
                    </div>
                    <div className="relative w-full">
                        <FloatingInput
                            ref={passwordRef}
                            id="password-textfield"
                            name="password"
                            label={`${translate("Password")} *`}
                            required
                            disabled={disabled}
                            value={password}
                            error={passwordError}
                            className="pr-10"
                            onChange={(v) => setPassword(v.target.value)}
                            onFocus={() => setPasswordError(false)}
                            type={showPassword ? "text" : "password"}
                            autoComplete={passkeyLogin ? "current-password webauthn" : "current-password"}
                            onKeyDown={handlePasswordKeyDown}
                            onKeyUp={handlePasswordKeyUp}
                        />
                        <PasswordVisibilityToggle
                            label={translate("Toggle password visibility")}
                            showPassword={showPassword}
                            {...toggleProps}
                        />
                    </div>
                    {passwordCapsLock ? (
                        <div className="w-full px-4">
                            <Alert variant="warning">
                                <AlertTitle>{translate("Warning")}</AlertTitle>
                                {passwordCapsLockPartial
                                    ? translate("The password was partially entered with Caps Lock")
                                    : translate("The password was entered with Caps Lock")}
                            </Alert>
                        </div>
                    ) : null}
                    {props.rememberMe ? (
                        <div className="-my-2 flex w-full flex-row">
                            <div className="flex flex-grow items-center gap-2">
                                <Checkbox
                                    id="remember-checkbox"
                                    disabled={disabled}
                                    checked={rememberMe}
                                    onCheckedChange={handleRememberMeChange}
                                    onKeyDown={handleRememberMeKeyDown}
                                />
                                <Label htmlFor="remember-checkbox" className="text-base">
                                    {translate("Remember me")}
                                </Label>
                            </div>
                        </div>
                    ) : null}
                    <div className="w-full">
                        <Button
                            id="sign-in-button"
                            type="submit"
                            variant="default"
                            className="w-full"
                            disabled={disabled}
                        >
                            {translate("Sign in")}
                            {isPending ? <Spinner size={20} className="ml-2 h-5 w-5" /> : null}
                        </Button>
                    </div>
                    {passkeyLogin ? (
                        <PasskeyForm
                            disabled={disabled}
                            rememberMe={props.rememberMe}
                            onAuthenticationError={(err) => createErrorNotification(err.message)}
                            onAuthenticationStart={() => {
                                setUsername("");
                                setPassword("");
                                setPasskeyAuthenticating(true);
                            }}
                            onAuthenticationStop={() => setPasskeyAuthenticating(false)}
                            onAuthenticationSuccess={props.onAuthenticationSuccess}
                        />
                    ) : null}
                    {props.resetPassword ? (
                        <div className="-my-2 flex w-full flex-row justify-end">
                            <button
                                id="reset-password-button"
                                type="button"
                                className="cursor-pointer py-[13.5px] text-base text-primary underline-offset-4 hover:underline"
                                onClick={handleResetPasswordClick}
                            >
                                {translate("Reset password?")}
                            </button>
                        </div>
                    ) : null}
                </div>
            </form>
        </LoginLayout>
    );
};

export default FirstFactorForm;
