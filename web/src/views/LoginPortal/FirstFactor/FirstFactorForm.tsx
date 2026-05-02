import { KeyboardEvent, useActionState, useEffect, useEffectEvent, useRef, useState } from "react";

import Visibility from "@mui/icons-material/Visibility";
import VisibilityOff from "@mui/icons-material/VisibilityOff";
import {
    Alert,
    AlertTitle,
    Button,
    Checkbox,
    CircularProgress,
    FormControlLabel,
    IconButton,
    InputAdornment,
    Link,
} from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { browserSupportsWebAuthn } from "@simplewebauthn/browser";
import { BroadcastChannel } from "broadcast-channel";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

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

const isWebAuthnSupported = browserSupportsWebAuthn();

export interface Props {
    passkeyLogin: boolean;
    rememberMe: boolean;
    resetPassword: boolean;
    resetPasswordCustomURL: string;

    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
    onChannelStateChange: () => void;
}

interface PasswordVisibilityToggleProps {
    visible: boolean;
    onVisibilityChange: (_visible: boolean) => void;
}

function PasswordVisibilityToggle({ onVisibilityChange, visible }: PasswordVisibilityToggleProps) {
    return (
        <InputAdornment position="end">
            <IconButton
                aria-label="toggle password visibility"
                edge="end"
                size="large"
                onMouseDown={() => onVisibilityChange(true)}
                onMouseUp={() => onVisibilityChange(false)}
                onMouseLeave={() => onVisibilityChange(false)}
                onTouchStart={() => onVisibilityChange(true)}
                onTouchEnd={() => onVisibilityChange(false)}
                onTouchCancel={() => onVisibilityChange(false)}
                onKeyDown={(e) => {
                    if (e.key === " " || e.key === "Enter") {
                        onVisibilityChange(true);
                        e.preventDefault();
                    }
                }}
                onKeyUp={(e) => {
                    if (e.key === " " || e.key === "Enter") {
                        onVisibilityChange(false);
                        e.preventDefault();
                    }
                }}
            >
                {visible ? <Visibility /> : <VisibilityOff />}
            </IconButton>
        </InputAdornment>
    );
}

export default function FirstFactorForm(props: Props) {
    const { t: translate } = useTranslation();

    const navigate = useNavigate();
    const redirectionURL = useQueryParam(RedirectionURL);
    const requestMethod = useQueryParam(RequestMethod);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();
    const { createErrorNotification } = useNotifications();

    const loginChannelRef = useRef<BroadcastChannel<boolean> | null>(null);
    const passwordRef = useRef<HTMLInputElement | null>(null);

    const [rememberMe, setRememberMe] = useState(false);
    const [username, setUsername] = useState("");
    const [usernameError, setUsernameError] = useState(false);
    const [password, setPassword] = useState("");
    const [showPassword, setShowPassword] = useState(false);
    const [passwordCapsLock, setPasswordCapsLock] = useState(false);
    const [passwordCapsLockPartial, setPasswordCapsLockPartial] = useState(false);
    const [passwordError, setPasswordError] = useState(false);
    const [passkeyAuthenticating, setPasskeyAuthenticating] = useState(false);

    const onChannelMessage = useEffectEvent((authenticated: boolean) => {
        if (authenticated) {
            props.onChannelStateChange();
        }
    });

    useEffect(() => {
        const channel = new BroadcastChannel<boolean>("login");
        loginChannelRef.current = channel;

        const handler = (authenticated: boolean) => onChannelMessage(authenticated);
        channel.addEventListener("message", handler);

        return () => {
            channel.removeEventListener("message", handler);
            void channel.close();
            loginChannelRef.current = null;
        };
    }, []);

    const handleRememberMeChange = () => {
        setRememberMe((prev) => !prev);
    };

    const [, signInAction, isPending] = useActionState<null>(async () => {
        if (username === "" || password === "") {
            if (username === "") setUsernameError(true);
            if (password === "") setPasswordError(true);
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
            passwordRef.current?.focus();
        }
        return null;
    }, null);

    const disabled = isPending || passkeyAuthenticating;

    const handleResetPasswordClick = () => {
        if (props.resetPassword) {
            if (props.resetPasswordCustomURL) {
                window.open(props.resetPasswordCustomURL);
            } else {
                navigate(ResetPasswordStep1Route);
            }
        }
    };

    const handlePasswordKeyUp = (event: KeyboardEvent<HTMLDivElement>) => {
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

    return (
        <LoginLayout id="first-factor-stage" title={translate("Sign in")}>
            <form id="form-login" action={signInAction} noValidate>
                <Grid container spacing={2}>
                    <Grid size={{ xs: 12 }}>
                        <TextField
                            id="username-textfield"
                            name="username"
                            label={translate("Username")}
                            variant="outlined"
                            required
                            autoFocus
                            value={username}
                            error={usernameError}
                            disabled={disabled}
                            fullWidth
                            onChange={(v) => setUsername(v.target.value)}
                            onFocus={() => setUsernameError(false)}
                            autoCapitalize="none"
                            autoComplete={props.passkeyLogin ? "username webauthn" : "username"}
                        />
                    </Grid>
                    <Grid size={{ xs: 12 }}>
                        <TextField
                            inputRef={passwordRef}
                            id="password-textfield"
                            name="password"
                            label={translate("Password")}
                            variant="outlined"
                            required
                            fullWidth
                            disabled={disabled}
                            value={password}
                            error={passwordError}
                            onChange={(v) => setPassword(v.target.value)}
                            onFocus={() => setPasswordError(false)}
                            type={showPassword ? "text" : "password"}
                            autoComplete={props.passkeyLogin ? "current-password webauthn" : "current-password"}
                            onKeyUp={handlePasswordKeyUp}
                            slotProps={{
                                input: {
                                    endAdornment: (
                                        <PasswordVisibilityToggle
                                            visible={showPassword}
                                            onVisibilityChange={setShowPassword}
                                        />
                                    ),
                                },
                            }}
                        />
                    </Grid>
                    {passwordCapsLock ? (
                        <Grid size={{ xs: 12 }} marginX={2}>
                            <Alert severity={"warning"}>
                                <AlertTitle>{translate("Warning")}</AlertTitle>
                                {passwordCapsLockPartial
                                    ? translate("The password was partially entered with Caps Lock")
                                    : translate("The password was entered with Caps Lock")}
                            </Alert>
                        </Grid>
                    ) : null}
                    {props.rememberMe ? (
                        <Grid
                            size={{ xs: 12 }}
                            sx={{
                                display: "flex",
                                flexDirection: "row",
                                marginBottom: (theme) => theme.spacing(-1),
                                marginTop: (theme) => theme.spacing(-1),
                            }}
                        >
                            <FormControlLabel
                                control={
                                    <Checkbox
                                        id="remember-checkbox"
                                        disabled={disabled}
                                        checked={rememberMe}
                                        onChange={handleRememberMeChange}
                                        value="rememberMe"
                                        color="primary"
                                    />
                                }
                                sx={{ flexGrow: 1 }}
                                label={translate("Remember me")}
                            />
                        </Grid>
                    ) : null}
                    <Grid size={{ xs: 12 }}>
                        <Button
                            id="sign-in-button"
                            type="submit"
                            variant="contained"
                            color="primary"
                            fullWidth={true}
                            endIcon={isPending ? <CircularProgress size={20} /> : null}
                            disabled={disabled}
                        >
                            {translate("Sign in")}
                        </Button>
                    </Grid>
                    {props.passkeyLogin && isWebAuthnSupported ? (
                        <PasskeyForm
                            disabled={disabled}
                            rememberMe={rememberMe}
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
                        <Grid
                            size={{ xs: 12 }}
                            sx={{
                                display: "flex",
                                flexDirection: "row",
                                justifyContent: "flex-end",
                                marginBottom: (theme) => theme.spacing(-1),
                                marginTop: (theme) => theme.spacing(-1),
                            }}
                        >
                            <Link
                                id="reset-password-button"
                                component="button"
                                onClick={handleResetPasswordClick}
                                sx={{ cursor: "pointer", paddingBottom: "13.5px", paddingTop: "13.5px" }}
                                underline="hover"
                            >
                                {translate("Reset password?")}
                            </Link>
                        </Grid>
                    ) : null}
                </Grid>
            </form>
        </LoginLayout>
    );
}
