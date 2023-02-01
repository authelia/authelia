import React, { Fragment, Suspense } from "react";

import { Box, Button, Paper, Stack, Typography } from "@mui/material";
import { useNavigate } from "react-router-dom";

import { RegisterWebauthnRoute } from "@constants/Routes";
import { AutheliaState } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import WebauthnDevicesStack from "@views/Settings/TwoFactorAuthentication/WebauthnDevicesStack";

interface Props {
    state: AutheliaState;
}

export default function WebauthnDevices(props: Props) {
    const navigate = useNavigate();

    const initiateRegistration = async (redirectRoute: string) => {
        navigate(redirectRoute);
    };

    const handleAddKeyButtonClick = () => {
        initiateRegistration(RegisterWebauthnRoute);
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
