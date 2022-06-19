import React, { MutableRefObject, useEffect, useRef, useState } from "react";

import { Grid, Button, FormControlLabel, Checkbox, Link, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import classnames from "classnames";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import FixedTextField from "@components/FixedTextField";
import { ResetPasswordStep1Route } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { usePageVisibility } from "@hooks/PageVisibility";
import { useRedirectionURL } from "@hooks/RedirectionURL";
import { useRequestMethod } from "@hooks/RequestMethod";
import { useAutheliaState } from "@hooks/State";
import LoginLayout from "@layouts/LoginLayout";
import { postFirstFactor } from "@services/FirstFactor";
import { AuthenticationLevel } from "@services/State";

export interface Props {
    disabled: boolean;
    rememberMe: boolean;

    resetPassword: boolean;
    resetPasswordCustomURL: string;

    onAuthenticationStart: () => void;
    onAuthenticationFailure: () => void;
    onAuthenticationSuccess: (redirectURL: string | undefined) => void;
}

const FirstFactorForm = function (props: Props) {
    const styles = useStyles();
    const navigate = useNavigate();
    const redirectionURL = useRedirectionURL();
    const requestMethod = useRequestMethod();

    const [state, fetchState, ,] = useAutheliaState();
    const [rememberMe, setRememberMe] = useState(false);
    const [username, setUsername] = useState("");
    const [usernameError, setUsernameError] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordError, setPasswordError] = useState(false);
    const { createErrorNotification } = useNotifications();
    // TODO (PR: #806, Issue: #511) potentially refactor
    const usernameRef = useRef() as MutableRefObject<HTMLInputElement>;
    const passwordRef = useRef() as MutableRefObject<HTMLInputElement>;
    const visible = usePageVisibility();
    const { t: translate } = useTranslation();

    useEffect(() => {
        const timeout = setTimeout(() => usernameRef.current.focus(), 10);
        return () => clearTimeout(timeout);
    }, [usernameRef]);

    useEffect(() => {
        if (visible) {
            fetchState();
        }
        const timer = setInterval(() => fetchState(), 1000);
        return () => clearInterval(timer);
    }, [visible, fetchState]);

    useEffect(() => {
        if (state && state.authentication_level >= AuthenticationLevel.OneFactor) {
            props.onAuthenticationSuccess(redirectionURL);
        }
    }, [state, redirectionURL, props]);

    const disabled = props.disabled;

    const handleRememberMeChange = () => {
        setRememberMe(!rememberMe);
    };

    const handleSignIn = async () => {
        if (username === "" || password === "") {
            if (username === "") {
                setUsernameError(true);
            }

            if (password === "") {
                setPasswordError(true);
            }
            return;
        }

        props.onAuthenticationStart();
        try {
            const res = await postFirstFactor(username, password, rememberMe, redirectionURL, requestMethod);
            props.onAuthenticationSuccess(res ? res.redirect : undefined);
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Incorrect username or password"));
            props.onAuthenticationFailure();
            setPassword("");
            passwordRef.current.focus();
        }
    };

    const handleResetPasswordClick = () => {
        if (props.resetPassword) {
            if (props.resetPasswordCustomURL !== "") {
                window.open(props.resetPasswordCustomURL);
            } else {
                navigate(ResetPasswordStep1Route);
            }
        }
    };

    return (
        <LoginLayout id="first-factor-stage" title={translate("Sign in")} showBrand>
            <Grid container spacing={2}>
                <Grid item xs={12}>
                    <FixedTextField
                        // TODO (PR: #806, Issue: #511) potentially refactor
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
                        onKeyPress={(ev) => {
                            if (ev.key === "Enter") {
                                if (!username.length) {
                                    setUsernameError(true);
                                } else if (username.length && password.length) {
                                    handleSignIn();
                                } else {
                                    setUsernameError(false);
                                    passwordRef.current.focus();
                                }
                            }
                        }}
                    />
                </Grid>
                <Grid item xs={12}>
                    <FixedTextField
                        // TODO (PR: #806, Issue: #511) potentially refactor
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
                        onKeyPress={(ev) => {
                            if (ev.key === "Enter") {
                                if (!username.length) {
                                    usernameRef.current.focus();
                                } else if (!password.length) {
                                    passwordRef.current.focus();
                                }
                                handleSignIn();
                                ev.preventDefault();
                            }
                        }}
                    />
                </Grid>
                {props.rememberMe ? (
                    <Grid item xs={12} className={classnames(styles.actionRow)}>
                        <FormControlLabel
                            control={
                                <Checkbox
                                    id="remember-checkbox"
                                    disabled={disabled}
                                    checked={rememberMe}
                                    onChange={handleRememberMeChange}
                                    onKeyPress={(ev) => {
                                        if (ev.key === "Enter") {
                                            if (!username.length) {
                                                usernameRef.current.focus();
                                            } else if (!password.length) {
                                                passwordRef.current.focus();
                                            }
                                            handleSignIn();
                                        }
                                    }}
                                    value="rememberMe"
                                    color="primary"
                                />
                            }
                            className={styles.rememberMe}
                            label={translate("Remember me")}
                        />
                    </Grid>
                ) : null}
                <Grid item xs={12}>
                    <Button
                        id="sign-in-button"
                        variant="contained"
                        color="primary"
                        fullWidth
                        disabled={disabled}
                        onClick={handleSignIn}
                    >
                        {translate("Sign in")}
                    </Button>
                </Grid>
                {props.resetPassword ? (
                    <Grid item xs={12} className={classnames(styles.actionRow, styles.flexEnd)}>
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
        </LoginLayout>
    );
};

export default FirstFactorForm;

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
