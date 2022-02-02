import React, { MutableRefObject, useState, useRef, useEffect, useCallback } from "react";

import { makeStyles, Grid, Button, Link } from "@material-ui/core";
import classnames from "classnames";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import FixedTextField from "@components/FixedTextField";
import { FirstFactorRoute } from "@constants/Routes";
import { useRedirectionURL } from "@hooks/RedirectionURL";
import { useRequestMethod } from "@hooks/RequestMethod";
import LoginLayout from "@layouts/LoginLayout";
import { makeParams } from "@utils/MakeParams";

export interface Props {
    signUp: boolean;
}

const SignUp: React.FC<Props> = function ({ signUp }) {
    const style = useStyles();
    const navigate = useNavigate();
    const redirectionURL = useRedirectionURL();
    const requestMethod = useRequestMethod();

    const [username, setUsername] = useState("");
    const [usernameError, setUsernameError] = useState(false);
    const [email, setEmail] = useState("");
    const [emailError, setEmailError] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordError, setPasswordError] = useState(false);

    // TODO (PR: #806, Issue: #511) potentially refactor
    const usernameRef = useRef() as MutableRefObject<HTMLInputElement>;
    const emailRef = useRef() as MutableRefObject<HTMLInputElement>;
    const passwordRef = useRef() as MutableRefObject<HTMLInputElement>;

    const { t: translate } = useTranslation("Portal");

    useEffect(() => {
        const timeout = setTimeout(() => usernameRef.current.focus(), 10);
        return () => clearTimeout(timeout);
    }, [usernameRef]);

    const handleSignUp = async () => {
        if (username === "" || email === "" || password === "") {
            if (username === "") {
                setUsernameError(true);
            }

            if (email === "") {
                setEmailError(true);
            }

            if (password === "") {
                setPasswordError(true);
            }
            return;
        }

        // TODO : call backend here
        // props.onAuthenticationStart();
        // try {
        //     const res = await postFirstFactor(username, email, password, redirectionURL, requestMethod);
        //     props.onAuthenticationSuccess(res ? res.redirect : undefined);
        // } catch (err) {
        //     console.error(err);
        //     createErrorNotification(translate("Incorrect username or password"));
        //     props.onAuthenticationFailure();
        //     setPassword("");
        //     passwordRef.current.focus();
        // }
    };

    const handleSignInClick = useCallback(() => {
        const params = makeParams({ rd: redirectionURL, rm: requestMethod });
        navigate(`${FirstFactorRoute}${params}`);
    }, [navigate, redirectionURL, requestMethod]);

    useEffect(() => {
        if (!signUp) {
            handleSignInClick();
        }
    }, [signUp, handleSignInClick]);

    return (
        <LoginLayout id="sign-up-stage" title={translate("Sign Up")} showBrand>
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
                                    handleSignUp();
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
                        inputRef={emailRef}
                        id="email-textfield"
                        label={translate("Email")}
                        variant="outlined"
                        required
                        value={email}
                        error={emailError}
                        fullWidth
                        onChange={(v) => setEmail(v.target.value)}
                        onFocus={() => setEmailError(false)}
                        autoCapitalize="none"
                        autoComplete="email"
                        onKeyPress={(ev) => {
                            if (ev.key === "Enter") {
                                if (!email.length) {
                                    setEmailError(true);
                                } else if (email.length && password.length) {
                                    handleSignUp();
                                } else {
                                    setEmailError(false);
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
                                handleSignUp();
                                ev.preventDefault();
                            }
                        }}
                    />
                </Grid>
                <Grid item xs={12}>
                    <Button id="sign-in-button" variant="contained" color="primary" fullWidth onClick={handleSignUp}>
                        {translate("Sign Up")}
                    </Button>
                </Grid>
                <Grid item xs={12} className={classnames(style.actionRow, style.center)}>
                    {translate("Already have an account?")}
                    &nbsp;
                    <Link id="signin-button" component="button" onClick={handleSignInClick}>
                        {translate("Sign In")}
                    </Link>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default SignUp;

const useStyles = makeStyles((theme) => ({
    actionRow: {
        display: "flex",
        flexDirection: "row",
        marginTop: theme.spacing(-1),
        marginBottom: theme.spacing(-1),
    },
    center: {
        justifyContent: "center",
    },
}));
