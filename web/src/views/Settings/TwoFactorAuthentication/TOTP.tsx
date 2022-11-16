import React, { useState } from "react";

import { Box, Button, Chip, Paper, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { RegisterOneTimePasswordRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { Configuration } from "@models/Configuration";
import { SecondFactorMethod } from "@models/Methods";
import { UserInfo } from "@models/UserInfo";
import { initiateTOTPRegistrationProcess } from "@root/services/RegisterDevice";
import { AutheliaState, AuthenticationLevel } from "@root/services/State";

interface Props {
    configuration: Configuration;
    state: AutheliaState;
    userInfo: UserInfo;
}

export default function TwoFactorAuthSettings(props: Props) {
    const { t: translate } = useTranslation("settings");
    const navigate = useNavigate();

    const { createInfoNotification, createErrorNotification } = useNotifications();
    const [registrationInProgress, setRegistrationInProgress] = useState(false);

    const initiateRegistration = async (initiateRegistrationFunc: () => Promise<void>, redirectRoute: string) => {
        if (props.state.authentication_level >= AuthenticationLevel.TwoFactor) {
            navigate(redirectRoute);
        } else {
            if (registrationInProgress) {
                return;
            }
            setRegistrationInProgress(true);
            try {
                await initiateRegistrationFunc();
                createInfoNotification(translate("An email has been sent to your address to complete the process"));
            } catch (err) {
                console.error(err);
                createErrorNotification(translate("There was a problem initiating the registration process"));
            }
            setRegistrationInProgress(false);
        }
    };

    const handleAddKeyButtonClick = () => {
        initiateRegistration(initiateTOTPRegistrationProcess, RegisterOneTimePasswordRoute);
    };

    return (
        <Paper variant="outlined">
            <Box sx={{ p: 3 }}>
                <Stack spacing={2}>
                    <Box>
                        <Typography variant="h5" style={{ flexGrow: 1 }}>
                            One-Time Password
                        </Typography>
                        <Stack direction="row" spacing={1}>
                            {!props.configuration.available_methods.has(SecondFactorMethod.TOTP) && (
                                <Chip label="not available" color="secondary" variant="outlined" />
                            )}
                            {props.userInfo.has_totp && (
                                <>
                                    <Chip label="enabled" color="primary" variant="outlined" />
                                    {props.userInfo.method === SecondFactorMethod.TOTP && (
                                        <Chip label="default" color="primary" />
                                    )}
                                </>
                            )}
                        </Stack>
                    </Box>
                    <Box>
                        <Button variant="outlined" color="primary" onClick={handleAddKeyButtonClick}>
                            {"Reset Device"}
                        </Button>
                    </Box>
                </Stack>
            </Box>
        </Paper>
    );
}
