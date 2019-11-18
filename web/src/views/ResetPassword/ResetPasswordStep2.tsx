import React, { useState, useCallback, useEffect } from "react";
import LoginLayout from "../../layouts/LoginLayout";
import classnames from "classnames";
import { Grid, Button, makeStyles } from "@material-ui/core";
import { useNotifications } from "../../hooks/NotificationsContext";
import { useHistory, useLocation } from "react-router";
import { completeResetPasswordProcess, resetPassword } from "../../services/ResetPassword";
import { FirstFactorRoute } from "../../Routes";
import { extractIdentityToken } from "../../utils/IdentityToken";
import FixedTextField from "../../components/FixedTextField";

export default function () {
    const style = useStyles();
    const location = useLocation();
    const [formDisabled, setFormDisabled] = useState(true);
    const [password1, setPassword1] = useState("");
    const [password2, setPassword2] = useState("");
    const [errorPassword1, setErrorPassword1] = useState(false);
    const [errorPassword2, setErrorPassword2] = useState(false);
    const { createSuccessNotification, createErrorNotification } = useNotifications();
    const history = useHistory();
    // Get the token from the query param to give it back to the API when requesting
    // the secret for OTP.
    const processToken = extractIdentityToken(location.search);

    const completeProcess = useCallback(async () => {
        if (!processToken) {
            setFormDisabled(true);
            createErrorNotification("No verification token provided");
            return;
        }

        try {
            setFormDisabled(true);
            await completeResetPasswordProcess(processToken);
            setFormDisabled(false);
        } catch (err) {
            console.error(err);
            createErrorNotification("There was an issue completing the process. " +
                "The verification token might have expired.");
            setFormDisabled(true);
        }
    }, [processToken, createErrorNotification]);

    useEffect(() => {
        completeProcess();
    }, [completeProcess]);

    const doResetPassword = async () => {
        if (password1 === "" || password2 === "") {
            if (password1 === "") {
                setErrorPassword1(true);
            }
            if (password2 === "") {
                setErrorPassword2(true);
            }

            if (password1 !== password2) {
                setErrorPassword1(true);
                setErrorPassword2(true)
            }
            return
        }

        try {
            await resetPassword(password1);
            createSuccessNotification("The password has been reset");
            setTimeout(() => history.push(FirstFactorRoute), 1500);
            setFormDisabled(true);
        } catch (err) {
            console.error(err);
            createErrorNotification("There was an issue resetting the password");
        }
    }

    const handleResetClick = () =>
        doResetPassword();

    const handleCancelClick = () =>
        history.push(FirstFactorRoute);

    return (
        <LoginLayout title="Enter new password">
            <Grid container className={style.root} spacing={2}>
                <Grid item xs={12}>
                    <FixedTextField
                        label="New password"
                        variant="outlined"
                        type="password"
                        value={password1}
                        disabled={formDisabled}
                        onChange={e => setPassword1(e.target.value)}
                        error={errorPassword1}
                        className={classnames(style.fullWidth)} />
                </Grid>
                <Grid item xs={12}>
                    <FixedTextField
                        label="Repeat new password"
                        variant="outlined"
                        type="password"
                        disabled={formDisabled}
                        value={password2}
                        onChange={e => setPassword2(e.target.value)}
                        error={errorPassword2}
                        onKeyPress={(ev) => {
                            if (ev.key === 'Enter') {
                                doResetPassword();
                                ev.preventDefault();
                            }
                        }}
                        className={classnames(style.fullWidth)} />
                </Grid>
                <Grid item xs={6}>
                    <Button
                        variant="contained"
                        color="primary"
                        name="password1"
                        disabled={formDisabled}
                        onClick={handleResetClick}
                        className={style.fullWidth}>Reset</Button>
                </Grid>
                <Grid item xs={6}>
                    <Button
                        variant="contained"
                        color="primary"
                        name="password2"
                        onClick={handleCancelClick}
                        className={style.fullWidth}>Cancel</Button>
                </Grid>
            </Grid>
        </LoginLayout>
    )
}

const useStyles = makeStyles(theme => ({
    root: {
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
    fullWidth: {
        width: "100%",
    }
}))