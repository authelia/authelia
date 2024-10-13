import React, { useState } from "react";

import { Button, FormControl, LinearProgress, Theme } from "@mui/material";
import Grid from "@mui/material/Grid2";
import TextField from "@mui/material/TextField";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { IndexRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import MinimalLayout from "@layouts/MinimalLayout";
import { initiateResetPasswordProcess } from "@services/ResetPassword";

const ResetPasswordStep1 = function () {
    const styles = useStyles();
    const [username, setUsername] = useState("");
    const [error, setError] = useState(false);
    const [loading, setLoading] = useState(false);
    const { createInfoNotification, createErrorNotification } = useNotifications();
    const navigate = useNavigate();
    const { t: translate } = useTranslation();

    const doInitiateResetPasswordProcess = async () => {
        if (username === "") {
            setError(true);
            setLoading(false);
            return;
        }

        try {
            await initiateResetPasswordProcess(username);
            createInfoNotification(translate("An email has been sent to your address to complete the process"));
        } catch {
            createErrorNotification(translate("There was an issue initiating the password reset process"));
        }
        setLoading(false);
    };

    const handleResetClick = () => {
        setLoading(true);
        doInitiateResetPasswordProcess();
    };

    const handleCancelClick = () => {
        setLoading(false);
        navigate(IndexRoute);
    };

    return (
        <MinimalLayout title={translate("Reset password")} id="reset-password-step1-stage">
            <FormControl id={"form-reset-password-username"}>
                <Grid container className={styles.root} spacing={2}>
                    <Grid size={{ xs: 12 }}>
                        <TextField
                            id="username-textfield"
                            label={translate("Username")}
                            disabled={loading}
                            variant="outlined"
                            fullWidth
                            error={error}
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            onKeyDown={(ev) => {
                                if (ev.key === "Enter") {
                                    setLoading(true);
                                    doInitiateResetPasswordProcess();
                                    ev.preventDefault();
                                }
                            }}
                        />
                        {loading ? <LinearProgress variant="indeterminate" sx={{ my: 1 }} /> : <></>}
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                        <Button
                            id="reset-button"
                            variant="contained"
                            disabled={loading}
                            color="primary"
                            fullWidth
                            onClick={handleResetClick}
                        >
                            {translate("Reset")}
                        </Button>
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                        <Button
                            id="cancel-button"
                            variant="contained"
                            disabled={loading}
                            color="primary"
                            fullWidth
                            onClick={handleCancelClick}
                        >
                            {translate("Cancel")}
                        </Button>
                    </Grid>
                </Grid>
            </FormControl>
        </MinimalLayout>
    );
};

export default ResetPasswordStep1;

const useStyles = makeStyles((theme: Theme) => ({
    root: {
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));
