import { KeyboardEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";

import { BroadcastChannel } from "broadcast-channel";
import { Eye, EyeOff } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";

import { Alert, AlertTitle } from "@components/UI/Alert";
import { Button } from "@components/UI/Button";
import { Checkbox } from "@components/UI/Checkbox";
import { FloatingInput } from "@components/UI/FloatingInput";
import { Label } from "@components/UI/Label";
import { Spinner } from "@components/UI/Spinner";
import { ResetPasswordStep1Route } from "@constants/Routes";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import LoginLayout from "@layouts/LoginLayout";
import { IsCapsLockModified } from "@services/CapsLock";
import { postFirstFactor } from "@services/Password";
import PasskeyForm from "@views/LoginPortal/FirstFactor/PasskeyForm";
export const isExpiredFlowError = (err: unknown) => {
    return err instanceof Error && err.message.includes("no longer valid");
};

export interface Props {
    disabled: boolean;
    passkeyLogin: boolean;
    rememberMe: boolean;
    resetPassword: boolean;
    resetPasswordCustomURL: string;

    onAuthenticationStart: () => void;
    onAuthenticationStop: () => void;
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

    const loginChannel = useMemo(() => new BroadcastChannel<boolean>("login"), []);

    const [rememberMe, setRememberMe] = useState(false);
    const [username, setUsername] = useState("");
    const [usernameError, setUsernameError] = useState(false);
    const [password, setPassword] = useState("");
    const [showPassword, setShowPassword] = useState(false);
    const [passwordCapsLock, setPasswordCapsLock] = useState(false);
    const [passwordCapsLockPartial, setPasswordCapsLockPartial] = useState(false);
    const [passwordError, setPasswordError] = useState(false);
    const [loading, setLoading] = useState(false);

    const usernameRef = useRef<HTMLInputElement | null>(null);
    const passwordRef = useRef<HTMLInputElement | null>(null);

    const focusUsername = useCallback(() => {
        if (usernameRef.current === null) return;

        usernameRef.current.focus();
    }, [usernameRef]);

    const focusPassword = useCallback(() => {
        if (passwordRef.current === null) return;

        passwordRef.current.focus();
    }, [passwordRef]);

    useEffect(() => {
        const timeout = setTimeout(() => focusUsername(), 10);
        return () => clearTimeout(timeout);
    }, [focusUsername]);

    useEffect(() => {
        const handleMessage = (authenticated: boolean) => {
            if (authenticated) {
                props.onChannelStateChange();
            }
        };

        loginChannel.addEventListener("message", handleMessage);

        return () => {
            loginChannel.removeEventListener("message", handleMessage);
        };
    }, [loginChannel, redirectionURL, props]);

    const disabled = props.disabled;

    const handleRememberMeChange = () => {
        setRememberMe(!rememberMe);
    };

    const handleSignIn = useCallback(async () => {
        if (loading) {
            return;
        }

        if (username === "" || password === "") {
            if (username === "") {
                setUsernameError(true);
            }

            if (password === "") {
                setPasswordError(true);
            }
            return;
        }

        setLoading(true);

        props.onAuthenticationStart();

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

            setLoading(false);

            await loginChannel.postMessage(true);
            props.onAuthenticationSuccess(res ? res.redirect : undefined);
        } catch (err) {
            console.error(err);

            if (isExpiredFlowError(err)) {
                // The submitted credentials were correct, but the sign-in flow is no
                // longer valid (e.g. the flow_id was already responded to or the
                // consent session expired). The authentication itself succeeded, so
                // informing the user their password was incorrect would be misleading.
                // Redirect to a clean sign-in page without flow parameters so they
                // can sign in again and avoid looping on the expired flow.
                const url = new URL(window.location.href);
                url.searchParams.delete("flow");
                url.searchParams.delete("flowID");
                url.searchParams.delete("subflow");

                setLoading(false);
                props.onAuthenticationStop();
                window.location.href = url.toString();

                return;
            }

            createErrorNotification(translate("Incorrect username or password"));
            setLoading(false);
            props.onAuthenticationStop();
            setPassword("");
            focusPassword();
        }
    }, [
        loading,
        username,
        password,
        props,
        rememberMe,
        redirectionURL,
        requestMethod,
        flowID,
        flow,
        subflow,
        userCode,
        loginChannel,
        createErrorNotification,
        translate,
        focusPassword,
    ]);

    const handleResetPasswordClick = () => {
        if (props.resetPassword) {
            if (props.resetPasswordCustomURL) {
                window.open(props.resetPasswordCustomURL);
            } else {
                navigate(ResetPasswordStep1Route);
            }
        }
    };

    const handleUsernameKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            if (event.key === "Enter") {
                if (!username.length) {
                    setUsernameError(true);
                } else if (username.length && password.length) {
                    handleSignIn().catch(console.error);
                } else {
                    setUsernameError(false);
                    focusPassword();
                }
                event.preventDefault();
            }
        },
        [focusPassword, handleSignIn, password.length, username.length],
    );

    const handlePasswordKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
            if (event.key === "Enter") {
                if (!username.length) {
                    focusUsername();
                } else if (!password.length) {
                    focusPassword();
                }
                handleSignIn().catch(console.error);
                event.preventDefault();
            }
        },
        [focusPassword, focusUsername, handleSignIn, password.length, username.length],
    );

    const handlePasswordKeyUp = useCallback(
        (event: KeyboardEvent<HTMLInputElement>) => {
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
        },
        [password.length],
    );

    const handleRememberMeKeyDown = useCallback(
        (event: KeyboardEvent<HTMLElement>) => {
            if (event.key === "Enter") {
                if (!username.length) {
                    focusUsername();
                } else if (!password.length) {
                    focusPassword();
                }
                handleSignIn().catch(console.error);
            }
        },
        [focusPassword, focusUsername, handleSignIn, password.length, username.length],
    );

    return (
        <LoginLayout id="first-factor-stage" title={translate("Sign in")}>
            <form id={"form-login"} onSubmit={(e) => e.preventDefault()}>
                <div className="grid grid-cols-1 gap-5">
                    <div className="w-full">
                        <FloatingInput
                            ref={usernameRef}
                            id="username-textfield"
                            label={`${translate("Username")} *`}
                            required
                            value={username}
                            error={usernameError}
                            disabled={disabled}
                            onChange={(v) => setUsername(v.target.value)}
                            onFocus={() => setUsernameError(false)}
                            autoCapitalize="none"
                            autoComplete="username"
                            onKeyDown={handleUsernameKeyDown}
                        />
                    </div>
                    <div className="relative w-full">
                        <FloatingInput
                            ref={passwordRef}
                            id="password-textfield"
                            label={`${translate("Password")} *`}
                            required
                            disabled={disabled}
                            value={password}
                            error={passwordError}
                            className="pr-10"
                            onChange={(v) => setPassword(v.target.value)}
                            onFocus={() => setPasswordError(false)}
                            type={showPassword ? "text" : "password"}
                            autoComplete="current-password"
                            onKeyDown={handlePasswordKeyDown}
                            onKeyUp={handlePasswordKeyUp}
                        />
                        <button
                            type="button"
                            className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
                            aria-label={translate("Toggle password visibility")}
                            aria-pressed={showPassword}
                            onMouseDown={() => setShowPassword(true)}
                            onMouseUp={() => setShowPassword(false)}
                            onMouseLeave={() => setShowPassword(false)}
                            onTouchStart={() => setShowPassword(true)}
                            onTouchEnd={() => setShowPassword(false)}
                            onTouchCancel={() => setShowPassword(false)}
                            onKeyDown={(e) => {
                                if (e.key === " " || e.key === "Enter") {
                                    setShowPassword(true);
                                    e.preventDefault();
                                }
                            }}
                            onKeyUp={(e) => {
                                if (e.key === " " || e.key === "Enter") {
                                    setShowPassword(false);
                                    e.preventDefault();
                                }
                            }}
                        >
                            {showPassword ? <Eye className="h-5 w-5" /> : <EyeOff className="h-5 w-5" />}
                        </button>
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
                            disabled={disabled || loading}
                            onClick={handleSignIn}
                        >
                            {translate("Sign in")}
                            {loading ? <Spinner size={20} className="ml-2 h-5 w-5" /> : null}
                        </Button>
                    </div>
                    {props.passkeyLogin ? (
                        <PasskeyForm
                            disabled={disabled || loading}
                            rememberMe={props.rememberMe}
                            onAuthenticationError={(err) => createErrorNotification(err.message)}
                            onAuthenticationStart={() => {
                                setUsername("");
                                setPassword("");
                                setLoading(true);
                                props.onAuthenticationStart();
                            }}
                            onAuthenticationStop={() => {
                                setLoading(false);
                                props.onAuthenticationStop();
                            }}
                            onAuthenticationSuccess={(url) => {
                                setLoading(false);
                                props.onAuthenticationSuccess(url);
                            }}
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
