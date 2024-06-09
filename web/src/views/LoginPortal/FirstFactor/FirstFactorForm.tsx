import React, { MutableRefObject, useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
    Alert,
    AlertTitle,
    Button,
    Checkbox,
    CircularProgress,
    FormControl,
    FormControlLabel,
    Link,
    Theme,
} from "@mui/material";
import Grid from "@mui/material/Grid2";
import TextField from "@mui/material/TextField";
import makeStyles from "@mui/styles/makeStyles";
import { BroadcastChannel } from "broadcast-channel";
import classnames from "classnames";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { ResetPasswordStep1Route } from "@constants/Routes";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useNotifications } from "@hooks/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useWorkflow } from "@hooks/Workflow";
import LoginLayout from "@layouts/LoginLayout";
import { IsCapsLockModified } from "@services/CapsLock";
import { postFirstFactor } from "@services/Password";
import PasskeyForm from "@views/LoginPortal/FirstFactor/PasskeyForm";

export interface Props {
    disabled: boolean;
    passkeyLogin: boolean;
    rememberMe: boolean;
    resetPassword: boolean;
    resetPasswordCustomURL: string;

    onAuthenticationStart: () => void;
    onAuthenticationStop: () => void;
    onAuthenticationSuccess: (redirectURL: string | undefined) => void;
    onChannelStateChange: () => void;
}

const FirstFactorForm = function (props: Props) {
    const { t: translate } = useTranslation();

    const navigate = useNavigate();
    const redirectionURL = useQueryParam(RedirectionURL);
    const requestMethod = useQueryParam(RequestMethod);
    const [workflow] = useWorkflow();
    const { createErrorNotification } = useNotifications();

    const loginChannel = useMemo(() => new BroadcastChannel<boolean>("login"), []);

    const [rememberMe, setRememberMe] = useState(false);
    const [username, setUsername] = useState("");
    const [usernameError, setUsernameError] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordCapsLock, setPasswordCapsLock] = useState(false);
    const [passwordCapsLockPartial, setPasswordCapsLockPartial] = useState(false);
    const [passwordError, setPasswordError] = useState(false);
    const [loading, setLoading] = useState(false);

    const usernameRef = useRef() as MutableRefObject<HTMLInputElement>;
    const passwordRef = useRef() as MutableRefObject<HTMLInputElement>;

    const styles = useStyles();

    useEffect(() => {
        const timeout = setTimeout(() => usernameRef.current.focus(), 10);
        return () => clearTimeout(timeout);
    }, [usernameRef]);

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
            const res = await postFirstFactor(username, password, rememberMe, redirectionURL, requestMethod, workflow);

            setLoading(false);

            await loginChannel.postMessage(true);
            props.onAuthenticationSuccess(res ? res.redirect : undefined);
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Incorrect username or password"));
            setLoading(false);
            props.onAuthenticationStop();
            setPassword("");
            passwordRef.current.focus();
        }
    }, [
        createErrorNotification,
        loginChannel,
        password,
        props,
        redirectionURL,
        rememberMe,
        requestMethod,
        translate,
        username,
        workflow,
    ]);

    const handleResetPasswordClick = () => {
        if (props.resetPassword) {
            if (props.resetPasswordCustomURL !== "") {
                window.open(props.resetPasswordCustomURL);
            } else {
                navigate(ResetPasswordStep1Route);
            }
        }
    };

    const handleUsernameKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                if (!username.length) {
                    setUsernameError(true);
                } else if (username.length && password.length) {
                    handleSignIn().catch(console.error);
                } else {
                    setUsernameError(false);
                    passwordRef.current.focus();
                }
            }
        },
        [handleSignIn, password.length, username.length],
    );

    const handlePasswordKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                if (!username.length) {
                    usernameRef.current.focus();
                } else if (!password.length) {
                    passwordRef.current.focus();
                }
                handleSignIn().catch(console.error);
                event.preventDefault();
            }
        },
        [handleSignIn, password.length, username.length],
    );

    const handlePasswordKeyUp = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
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
        (event: React.KeyboardEvent<HTMLButtonElement>) => {
            if (event.key === "Enter") {
                if (!username.length) {
                    usernameRef.current.focus();
                } else if (!password.length) {
                    passwordRef.current.focus();
                }
                handleSignIn().catch(console.error);
            }
        },
        [handleSignIn, password.length, username.length],
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
                            type="password"
                            autoComplete="current-password"
                            onKeyDown={handlePasswordKeyDown}
                            onKeyUp={handlePasswordKeyUp}
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
                        <Grid size={{ xs: 12 }} className={classnames(styles.actionRow)}>
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
                                className={styles.rememberMe}
                                label={translate("Remember me")}
                            />
                        </Grid>
                    ) : null}
                    <Grid size={{ xs: 12 }}>
                        <Button
                            id="sign-in-button"
                            variant="contained"
                            color="primary"
                            fullWidth
                            disabled={disabled}
                            onClick={handleSignIn}
                            endIcon={loading ? <CircularProgress size={20} /> : null}
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
                        <Grid size={{ xs: 12 }} className={classnames(styles.actionRow, styles.flexEnd)}>
                            <Link
                                id="reset-password-button"
                                component="button"
                                onClick={handleResetPasswordClick}
                                className={styles.resetLink}
                                underline="hover"
                            >
                                {translate("Reset password?")}
                            </Link>
                        </Grid>
                    ) : null}
                </Grid>
            </FormControl>
        </LoginLayout>
    );
};

const useStyles = makeStyles((theme: Theme) => ({
    actionRow: {
        display: "flex",
        flexDirection: "row",
        marginTop: theme.spacing(-1),
        marginBottom: theme.spacing(-1),
    },
    resetLink: {
        cursor: "pointer",
        paddingTop: 13.5,
        paddingBottom: 13.5,
    },
    rememberMe: {
        flexGrow: 1,
    },
    flexEnd: {
        justifyContent: "flex-end",
    },
}));

export default FirstFactorForm;
