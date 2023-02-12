import React, { useState } from "react";

import { Button, Grid, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import FixedTextField from "@components/FixedTextField";
import { IndexRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import LoginLayout from "@layouts/LoginLayout";
import { initiateResetPasswordProcess } from "@services/ResetPassword";

const ResetPasswordStep1 = function () {
    const styles = useStyles();
    const [username, setUsername] = useState("");
    const [error, setError] = useState(false);
    const { createInfoNotification, createErrorNotification } = useNotifications();
    const navigate = useNavigate();
    const { t: translate } = useTranslation();

    const doInitiateResetPasswordProcess = async () => {
        if (username === "") {
            setError(true);
            return;
        }

        try {
            await initiateResetPasswordProcess(username);
            createInfoNotification(translate("An email has been sent to your address to complete the process"));
        } catch (err) {
            createErrorNotification(translate("There was an issue initiating the password reset process"));
        }
    };

    const handleResetClick = () => {
        doInitiateResetPasswordProcess();
    };

    const handleCancelClick = () => {
        navigate(IndexRoute);
    };

    return (
        <LoginLayout title={translate("Reset password")} id="reset-password-step1-stage">
            <Grid container className={styles.root} spacing={2}>
                <Grid item xs={12}>
                    <FixedTextField
                        id="username-textfield"
                        label={translate("Username")}
                        variant="outlined"
                        fullWidth
                        error={error}
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        onKeyDown={(ev) => {
                            if (ev.key === "Enter") {
                                doInitiateResetPasswordProcess();
                                ev.preventDefault();
                            }
                        }}
                    />
                </Grid>
                <Grid item xs={6}>
                    <Button id="reset-button" variant="contained" color="primary" fullWidth onClick={handleResetClick}>
                        {translate("Reset")}
                    </Button>
                </Grid>
                <Grid item xs={6}>
                    <Button
                        id="cancel-button"
                        variant="contained"
                        color="primary"
                        fullWidth
                        onClick={handleCancelClick}
                    >
                        {translate("Cancel")}
                    </Button>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default ResetPasswordStep1;

const useStyles = makeStyles((theme: Theme) => ({
    root: {
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));
