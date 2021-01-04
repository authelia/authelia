import React, { useState } from "react";

import { Grid, Button, makeStyles } from "@material-ui/core";
import { useHistory } from "react-router";

import FixedTextField from "../../components/FixedTextField";
import { useNotifications } from "../../hooks/NotificationsContext";
import LoginLayout from "../../layouts/LoginLayout";
import { FirstFactorRoute } from "../../Routes";
import { initiateResetPasswordProcess } from "../../services/ResetPassword";

const ResetPasswordStep1 = function () {
    const style = useStyles();
    const [username, setUsername] = useState("");
    const [error, setError] = useState(false);
    const { createInfoNotification, createErrorNotification } = useNotifications();
    const history = useHistory();

    const doInitiateResetPasswordProcess = async () => {
        if (username === "") {
            setError(true);
            return;
        }

        try {
            await initiateResetPasswordProcess(username);
            createInfoNotification("An email has been sent to your address to complete the process.");
        } catch (err) {
            createErrorNotification("There was an issue initiating the password reset process.");
        }
    };

    const handleResetClick = () => {
        doInitiateResetPasswordProcess();
    };

    const handleCancelClick = () => {
        history.push(FirstFactorRoute);
    };

    return (
        <LoginLayout title="Reset password" id="reset-password-step1-stage">
            <Grid container className={style.root} spacing={2}>
                <Grid item xs={12}>
                    <FixedTextField
                        id="username-textfield"
                        label="Username"
                        variant="outlined"
                        fullWidth
                        error={error}
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        onKeyPress={(ev) => {
                            if (ev.key === "Enter") {
                                doInitiateResetPasswordProcess();
                                ev.preventDefault();
                            }
                        }}
                    />
                </Grid>
                <Grid item xs={6}>
                    <Button id="reset-button" variant="contained" color="primary" fullWidth onClick={handleResetClick}>
                        Reset
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
                        Cancel
                    </Button>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default ResetPasswordStep1;

const useStyles = makeStyles((theme) => ({
    root: {
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));
