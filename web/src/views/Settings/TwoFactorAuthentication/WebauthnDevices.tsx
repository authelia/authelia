import React, { Fragment, Suspense, useState } from "react";

import { Box, Button, Paper, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { RegisterWebauthnRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { initiateWebauthnRegistrationProcess } from "@services/RegisterDevice";
import { AutheliaState, AuthenticationLevel } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import WebauthnDevicesStack from "@views/Settings/TwoFactorAuthentication/WebauthnDevicesStack";

interface Props {
    state: AutheliaState;
}

export default function WebauthnDevices(props: Props) {
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
        initiateRegistration(initiateWebauthnRegistrationProcess, RegisterWebauthnRoute);
    };

    return (
        <Fragment>
            <Paper variant="outlined">
                <Box sx={{ p: 3 }}>
                    <Stack spacing={2}>
                        <Box>
                            <Typography variant="h5">Webauthn Devices</Typography>
                        </Box>
                        <Box>
                            <Button variant="outlined" color="primary" onClick={handleAddKeyButtonClick}>
                                {"Add new device"}
                            </Button>
                        </Box>
                        <Suspense fallback={<LoadingPage />}>
                            <WebauthnDevicesStack />
                        </Suspense>
                    </Stack>
                </Box>
            </Paper>
        </Fragment>
    );
}
