import React, { Fragment, Suspense, useState } from "react";

import { Box, Button, Paper, Stack, Tooltip, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import { AutheliaState } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";
import WebAuthnDeviceRegisterDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnDeviceRegisterDialog";
import WebAuthnDevicesStack from "@views/Settings/TwoFactorAuthentication/WebAuthnDevicesStack";

interface Props {
    state: AutheliaState;
}

export default function WebAuthnDevices(props: Props) {
    const { t: translate } = useTranslation("settings");

    const [showWebAuthnDeviceRegisterDialog, setShowWebAuthnDeviceRegisterDialog] = useState<boolean>(false);
    const [refreshState, setRefreshState] = useState<number>(0);

    const handleIncrementRefreshState = () => {
        setRefreshState((refreshState) => refreshState + 1);
    };

    return (
        <Fragment>
            <WebAuthnDeviceRegisterDialog
                open={showWebAuthnDeviceRegisterDialog}
                onClose={() => {
                    handleIncrementRefreshState();
                }}
                setCancelled={() => {
                    setShowWebAuthnDeviceRegisterDialog(false);
                    handleIncrementRefreshState();
                }}
            />
            <Paper variant="outlined">
                <Box sx={{ p: 3 }}>
                    <Stack spacing={2}>
                        <Box>
                            <Typography variant="h5">{translate("WebAuthn Credentials")}</Typography>
                        </Box>
                        <Box>
                            <Tooltip title={translate("Click to add a WebAuthn credential to your account")}>
                                <Button
                                    variant="outlined"
                                    color="primary"
                                    onClick={() => {
                                        setShowWebAuthnDeviceRegisterDialog(true);
                                    }}
                                >
                                    {translate("Add Credential")}
                                </Button>
                            </Tooltip>
                        </Box>
                        <Suspense fallback={<LoadingPage />}>
                            <WebAuthnDevicesStack
                                refreshState={refreshState}
                                incrementRefreshState={handleIncrementRefreshState}
                            />
                        </Suspense>
                    </Stack>
                </Box>
            </Paper>
        </Fragment>
    );
}
