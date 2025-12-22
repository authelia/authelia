import { KeyboardEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";

import Visibility from "@mui/icons-material/Visibility";
import VisibilityOff from "@mui/icons-material/VisibilityOff";
import {
    Alert,
    AlertTitle,
    Button,
    Checkbox,
    CircularProgress,
    Divider,
    FormControl,
    FormControlLabel,
    IconButton,
    InputAdornment,
    Link,
    Theme,
} from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { BroadcastChannel } from "broadcast-channel";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { makeStyles } from "tss-react/mui";

import { ResetPasswordStep1Route } from "@constants/Routes";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import LoginLayout from "@layouts/LoginLayout";
import { spnegoPostFirstFactor } from "@root/services/Spnego";
import { IsCapsLockModified } from "@services/CapsLock";
import { postFirstFactor } from "@services/Password";
import PasskeyForm from "@views/LoginPortal/FirstFactor/PasskeyForm";

export interface Props {
    disabled: boolean;
    passkeyLogin: boolean;
    rememberMe: boolean;
    resetPassword: boolean;
    resetPasswordCustomURL: string;
    spnegoLogin: boolean;

    onAuthenticationStart: () => void;
    onAuthenticationStop: () => void;
    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
    onChannelStateChange: () => void;
}

const FirstFactorForm = function (props: Props) {
    const { t: translate } = useTranslation();
    const { classes, cx } = useStyles();

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
        loginChannel.addEventListener("message", (authenticated) => {
            if (authenticated) {
                props.onChannelStateChange();
            }
        });
    }, [loginChannel, redirectionURL, props]);

    const disabled = props.disabled;

    const handleRememberMeChange = () => {
        setRememberMe(!rememberMe);
    };

    const handleSignIn = useCallback(async () => {
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
            createErrorNotification(translate("Incorrect username or password"));
            setLoading(false);
            props.onAuthenticationStop();
            setPassword("");
            focusPassword();
        }
    }, [
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

    const spnegoLogin = useCallback(async () => {
        setLoading(true);

        props.onAuthenticationStart();

        try {
            const res = await spnegoPostFirstFactor(
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
            createErrorNotification(translate("SPNEGO authentication failed"));
            setLoading(false);
            props.onAuthenticationStop();
            setPassword("");
            focusPassword();
        }
    }, [
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
        (event: KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                if (!username.length) {
                    setUsernameError(true);
                } else if (username.length && password.length) {
                    handleSignIn().catch(console.error);
                } else {
                    setUsernameError(false);
                    focusPassword();
                }
            }
        },
        [focusPassword, handleSignIn, password.length, username.length],
    );

    const handlePasswordKeyDown = useCallback(
        (event: KeyboardEvent<HTMLDivElement>) => {
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
        (event: KeyboardEvent<HTMLDivElement>) => {
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
        (event: KeyboardEvent<HTMLButtonElement>) => {
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
            <FormControl id={"form-login"}>
                <Grid container spacing={2}>
                    <Grid size={{ xs: 12 }}>
                        <TextField
                            inputRef={usernameRef}
                            id="username-textfield"
                            label={translate("Username")}
                            variant="outlined"
                            required
                            value={username}
                            error={usernameError}
                            disabled={disabled}
                            fullWidth
                            onChange={(v) => setUsername(v.target.value)}
                            onFocus={() => setUsernameError(false)}
                            autoCapitalize="none"
                            autoComplete="username"
                            onKeyDown={handleUsernameKeyDown}
                        />
                    </Grid>
                    <Grid size={{ xs: 12 }}>
                        <TextField
                            inputRef={passwordRef}
                            id="password-textfield"
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
                            autoComplete="current-password"
                            onKeyDown={handlePasswordKeyDown}
                            onKeyUp={handlePasswordKeyUp}
                            slotProps={{
                                input: {
                                    endAdornment: (
                                        <InputAdornment position="end">
                                            <IconButton
                                                aria-label="toggle password visibility"
                                                edge="end"
                                                size="large"
                                                onMouseDown={() => setShowPassword(true)}
                                                onMouseUp={() => setShowPassword(false)}
                                                onMouseLeave={() => setShowPassword(false)}
                                                onTouchStart={() => setShowPassword(true)}
                                                onTouchEnd={() => setShowPassword(false)}
                                                onTouchCancel={() => setShowPassword(false)}
                                                onKeyDown={(e) => {
                                                    if (e.key === " ") {
                                                        setShowPassword(true);
                                                        e.preventDefault();
                                                    }
                                                }}
                                                onKeyUp={(e) => {
                                                    if (e.key === " ") {
                                                        setShowPassword(false);
                                                        e.preventDefault();
                                                    }
                                                }}
                                            >
                                                {showPassword ? <Visibility /> : <VisibilityOff />}
                                            </IconButton>
                                        </InputAdornment>
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
                        <Grid size={{ xs: 12 }} className={cx(classes.actionRow)}>
                            <FormControlLabel
                                control={
                                    <Checkbox
                                        id="remember-checkbox"
                                        disabled={disabled}
                                        checked={rememberMe}
                                        onChange={handleRememberMeChange}
                                        onKeyDown={handleRememberMeKeyDown}
                                        value="rememberMe"
                                        color="primary"
                                    />
                                }
                                className={classes.rememberMe}
                                label={translate("Remember me")}
                            />
                        </Grid>
                    ) : null}
                    <Grid size={{ xs: 12 }}>
                        <Button
                            id="sign-in-button"
                            variant="contained"
                            color="primary"
                            fullWidth={true}
                            endIcon={loading ? <CircularProgress size={20} /> : null}
                            disabled={disabled}
                            onClick={handleSignIn}
                        >
                            {translate("Sign in")}
                        </Button>
                    </Grid>

                    {props.passkeyLogin ? (
                        <PasskeyForm
                            disabled={props.disabled}
                            rememberMe={props.rememberMe}
                            onAuthenticationError={(err) => createErrorNotification(err.message)}
                            onAuthenticationStart={() => {
                                setUsername("");
                                setPassword("");
                                props.onAuthenticationStart();
                            }}
                            onAuthenticationStop={props.onAuthenticationStop}
                            onAuthenticationSuccess={props.onAuthenticationSuccess}
                        />
                    ) : null}
                    {props.resetPassword ? (
                        <Grid size={{ xs: 12 }} className={cx(classes.actionRow, classes.flexEnd)}>
                            <Link
                                id="reset-password-button"
                                component="button"
                                onClick={handleResetPasswordClick}
                                className={classes.resetLink}
                                underline="hover"
                            >
                                {translate("Reset password?")}
                            </Link>
                        </Grid>
                    ) : null}

                    {props.spnegoLogin ? (
                        <Grid size={{ xs: 12 }}>
                            <Divider
                                orientation="horizontal"
                                variant="middle"
                                flexItem
                                style={{
                                    marginBottom: 16,
                                }}
                            />

                            <Button
                                id="spnego-login-button"
                                variant="contained"
                                color="primary"
                                fullWidth={true}
                                endIcon={loading ? <CircularProgress size={20} /> : null}
                                disabled={disabled}
                                onClick={spnegoLogin}
                            >
                                {translate("Login with SPNEGO")}
                            </Button>
                        </Grid>
                    ) : null}
                </Grid>
            </FormControl>
        </LoginLayout>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    actionRow: {
        display: "flex",
        flexDirection: "row",
        marginBottom: theme.spacing(-1),
        marginTop: theme.spacing(-1),
    },
    flexEnd: {
        justifyContent: "flex-end",
    },
    rememberMe: {
        flexGrow: 1,
    },
    resetLink: {
        cursor: "pointer",
        paddingBottom: 13.5,
        paddingTop: 13.5,
    },
}));

export default FirstFactorForm;
