import React, { MutableRefObject, useEffect, useRef, useState } from "react";

import { makeStyles, Grid, Button, FormControlLabel, Checkbox, Link } from "@material-ui/core";
import classnames from "classnames";
import { useHistory } from "react-router";

import FixedTextField from "../../../components/FixedTextField";
import { useNotifications } from "../../../hooks/NotificationsContext";
import { useRedirectionURL } from "../../../hooks/RedirectionURL";
import LoginLayout from "../../../layouts/LoginLayout";
import { ResetPasswordStep1Route } from "../../../Routes";
import { postFirstFactor } from "../../../services/FirstFactor";

export interface Props {
    disabled: boolean;
    rememberMe: boolean;
    resetPassword: boolean;

    onAuthenticationStart: () => void;
    onAuthenticationFailure: () => void;
    onAuthenticationSuccess: (redirectURL: string | undefined) => void;
}

const FirstFactorForm = function (props: Props) {
    const style = useStyles();
    const history = useHistory();
    const redirectionURL = useRedirectionURL();

    const [rememberMe, setRememberMe] = useState(false);
    const [username, setUsername] = useState("");
    const [usernameError, setUsernameError] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordError, setPasswordError] = useState(false);
    const { createErrorNotification } = useNotifications();
    // TODO (PR: #806, Issue: #511) potentially refactor
    const usernameRef = useRef() as MutableRefObject<HTMLInputElement>;
    const passwordRef = useRef() as MutableRefObject<HTMLInputElement>;
    useEffect(() => {
        const timeout = setTimeout(() => usernameRef.current.focus(), 10);
        return () => clearTimeout(timeout);
    }, [usernameRef]);

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
            const res = await postFirstFactor(username, password, rememberMe, redirectionURL);
            props.onAuthenticationSuccess(res ? res.redirect : undefined);
        } catch (err) {
            console.error(err);
            createErrorNotification("Incorrect username or password.");
            props.onAuthenticationFailure();
            setPassword("");
            passwordRef.current.focus();
        }
    };

    const handleResetPasswordClick = () => {
        history.push(ResetPasswordStep1Route);
    };

    return (
        <LoginLayout id="first-factor-stage" title="Sign in" showBrand>
            <Grid container spacing={2} className={style.root}>
                <Grid item xs={12}>
                    <FixedTextField
                        // TODO (PR: #806, Issue: #511) potentially refactor
                        inputRef={usernameRef}
                        id="username-textfield"
                        label="Username"
                        variant="outlined"
                        required
                        value={username}
                        error={usernameError}
                        disabled={disabled}
                        fullWidth
                        onChange={(v) => setUsername(v.target.value)}
                        onFocus={() => setUsernameError(false)}
                        autoCapitalize="none"
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
                        label="Password"
                        variant="outlined"
                        required
                        fullWidth
                        disabled={disabled}
                        value={password}
                        error={passwordError}
                        onChange={(v) => setPassword(v.target.value)}
                        onFocus={() => setPasswordError(false)}
                        type="password"
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
                {props.rememberMe || props.resetPassword ? (
                    <Grid
                        item
                        xs={12}
                        className={
                            props.rememberMe
                                ? classnames(style.leftAlign, style.actionRow)
                                : classnames(style.leftAlign, style.flexEnd, style.actionRow)
                        }
                    >
                        {props.rememberMe ? (
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
                                className={style.rememberMe}
                                label="Remember me"
                            />
                        ) : null}
                        {props.resetPassword ? (
                            <Link
                                id="reset-password-button"
                                component="button"
                                onClick={handleResetPasswordClick}
                                className={style.resetLink}
                            >
                                Reset password?
                            </Link>
                        ) : null}
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
                        Sign in
                    </Button>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default FirstFactorForm;

const useStyles = makeStyles((theme) => ({
    root: {
        marginTop: theme.spacing(),
        marginBottom: theme.spacing(),
    },
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
    leftAlign: {
        textAlign: "left",
    },
    rightAlign: {
        textAlign: "right",
        verticalAlign: "bottom",
    },
}));
