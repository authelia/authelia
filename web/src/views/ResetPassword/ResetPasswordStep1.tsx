import React, { useCallback, useEffect, useRef, useState } from "react";

import { Button, CircularProgress, FormControl, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import ComponentWithTooltip from "@components/ComponentWithTooltip";
import { IndexRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import MinimalLayout from "@layouts/MinimalLayout";
import { initiateResetPasswordProcess } from "@services/ResetPassword";

const ResetPasswordStep1 = function () {
    const theme = useTheme();
    const [username, setUsername] = useState("");
    const [error, setError] = useState(false);
    const [loading, setLoading] = useState(false);

    const [rateLimited, setRateLimited] = useState(false);
    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);

    const { createInfoNotification, createErrorNotification } = useNotifications();
    const navigate = useNavigate();
    const { t: translate } = useTranslation();

    useEffect(() => {
        return () => {
            if (timeoutRateLimit.current !== null) {
                clearTimeout(timeoutRateLimit.current);
                timeoutRateLimit.current = null;
            }
        };
    }, []);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            setRateLimited(true);

            createErrorNotification(translate("You have made too many requests"));

            timeoutRateLimit.current = setTimeout(() => {
                setRateLimited(false);
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [createErrorNotification, translate],
    );

    const doInitiateResetPasswordProcess = async () => {
        setError(false);
        setLoading(true);

        if (username === "") {
            setError(true);
            setLoading(false);
            createErrorNotification(translate("Username is required"));
            return;
        }

        try {
            const response = await initiateResetPasswordProcess(username);
            if (response && !response.limited) {
                createInfoNotification(translate("An email has been sent to your address to complete the process"));
                navigate(IndexRoute);
            } else if (response && response.limited) {
                handleRateLimited(response.retryAfter);
            } else {
                createErrorNotification(translate("There was an issue initiating the password reset process"));
            }
        } catch {
            createErrorNotification(translate("There was an issue initiating the password reset process"));
        }
        setLoading(false);
    };

    const handleResetClick = () => {
        doInitiateResetPasswordProcess();
    };

    const handleCancelClick = () => {
        navigate(IndexRoute);
    };

    return (
        <MinimalLayout title={translate("Reset password")} id="reset-password-step1-stage">
            <FormControl id={"form-reset-password-username"}>
                <Grid container sx={{ marginY: theme.spacing(2) }} spacing={2}>
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
                                    ev.preventDefault();
                                    doInitiateResetPasswordProcess();
                                }
                            }}
                        />
                    </Grid>
                    <Grid size={{ xs: 6 }}>
                        <ComponentWithTooltip render={rateLimited} title={translate("You have made too many requests")}>
                            <Button
                                id="reset-button"
                                variant="contained"
                                disabled={loading || rateLimited}
                                color="primary"
                                fullWidth
                                onClick={handleResetClick}
                                startIcon={loading ? <CircularProgress color="inherit" size={20} /> : <></>}
                            >
                                {translate("Reset")}
                            </Button>
                        </ComponentWithTooltip>
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
