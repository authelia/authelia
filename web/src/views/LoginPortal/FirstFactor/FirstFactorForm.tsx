import React, { useState } from "react";
import classnames from "classnames";
import { makeStyles, Grid, Button, FormControlLabel, Checkbox, Link } from "@material-ui/core";
import { useHistory } from "react-router";
import LoginLayout from "../../../layouts/LoginLayout";
import { useNotifications } from "../../../hooks/NotificationsContext";
import { postFirstFactor } from "../../../services/FirstFactor";
import { ResetPasswordStep1Route } from "../../../Routes";
import { useRedirectionURL } from "../../../hooks/RedirectionURL";
import FixedTextField from "../../../components/FixedTextField";
import {useRequestMethod} from "../../../hooks/RequestMethod";

export interface Props {
    disabled: boolean;

    onAuthenticationStart: () => void;
    onAuthenticationFailure: () => void;
    onAuthenticationSuccess: (redirectURL: string | undefined) => void;
}

export default function (props: Props) {
    const style = useStyles();
    const history = useHistory();
    const redirectionURL = useRedirectionURL();
    const requestMethod = useRequestMethod();

    const [rememberMe, setRememberMe] = useState(false);
    const [username, setUsername] = useState("");
    const [usernameError, setUsernameError] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordError, setPasswordError] = useState(false);
    const { createErrorNotification } = useNotifications();

    const disabled = props.disabled;

    const handleRememberMeChange = () => {
        setRememberMe(!rememberMe);
    }

    const handleSignIn = async () => {
        if (username === "" || password === "") {
            if (username === "") {
                setUsernameError(true)
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
            createErrorNotification(
                "There was a problem. Username or password might be incorrect.");
            props.onAuthenticationFailure();
        }
    }

    const handleResetPasswordClick = () => {
        history.push(ResetPasswordStep1Route);
    }

    return (
        <LoginLayout
            id="first-factor-stage"
            title="Sign in"
            showBrand>
            <Grid container spacing={2} className={style.root}>
                <Grid item xs={12}>
                    <FixedTextField
                        id="username-textfield"
                        label="Username"
                        variant="outlined"
                        required
                        value={username}
                        error={usernameError}
                        disabled={disabled}
                        fullWidth
                        onChange={v => setUsername(v.target.value)}
                        onFocus={() => setUsernameError(false)} />
                </Grid>
                <Grid item xs={12}>
                    <FixedTextField
                        id="password-textfield"
                        label="Password"
                        variant="outlined"
                        required
                        fullWidth
                        disabled={disabled}
                        value={password}
                        error={passwordError}
                        onChange={v => setPassword(v.target.value)}
                        onFocus={() => setPasswordError(false)}
                        type="password"
                        onKeyPress={(ev) => {
                            if (ev.key === 'Enter') {
                                handleSignIn();
                                ev.preventDefault();
                            }
                        }} />
                </Grid>
                <Grid item xs={12} className={classnames(style.leftAlign, style.actionRow)}>
                    <FormControlLabel
                        control={
                            <Checkbox
                                id="remember-checkbox"
                                disabled={disabled}
                                checked={rememberMe}
                                onChange={handleRememberMeChange}
                                value="rememberMe"
                                color="primary" />
                        }
                        className={style.rememberMe}
                        label="Remember me"
                    />
                    <Link
                        id="reset-password-button"
                        component="button"
                        onClick={handleResetPasswordClick}
                        className={style.resetLink}>
                        Reset password?
                    </Link>
                </Grid>
                <Grid item xs={12}>
                    <Button
                        id="sign-in-button"
                        variant="contained"
                        color="primary"
                        fullWidth
                        disabled={disabled}
                        onClick={handleSignIn}>
                        Sign in
                    </Button>
                </Grid>
            </Grid>
        </LoginLayout>
    )
}

const useStyles = makeStyles(theme => ({
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
    },
    rememberMe: {
        flexGrow: 1,
    },
    leftAlign: {
        textAlign: "left",
    },
    rightAlign: {
        textAlign: "right",
        verticalAlign: "bottom",
    },
}))