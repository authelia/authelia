import React, { Fragment, useState } from "react";

import { Button, Paper, Tooltip, Typography } from "@mui/material";
import Grid from "@mui/material/Unstable_Grid2";
import { useTranslation } from "react-i18next";

import { WebAuthnDevice } from "@models/WebAuthn";
import WebAuthnDeviceRegisterDialog from "@views/Settings/TwoFactorAuthentication/WebAuthnDeviceRegisterDialog";
import WebAuthnDevicesStack from "@views/Settings/TwoFactorAuthentication/WebAuthnDevicesStack";

interface Props {
    devices: WebAuthnDevice[] | undefined;
    handleRefreshState: () => void;
}

export default function WebAuthnDevicesPanel(props: Props) {
    const { t: translate } = useTranslation("settings");

    const [showRegisterDialog, setShowRegisterDialog] = useState<boolean>(false);

    return (
        <Fragment>
            <WebAuthnDeviceRegisterDialog
                open={showRegisterDialog}
                setCancelled={() => {
                    setShowRegisterDialog(false);
                    props.handleRefreshState();
                }}
            />
            <Paper variant={"outlined"}>
                <Grid container spacing={2} padding={2}>
                    <Grid xs={12}>
                        <Typography variant="h5">{translate("WebAuthn Credentials")}</Typography>
                    </Grid>
                    <Grid xs={4} md={2}>
                        <Tooltip title={translate("Click to add a WebAuthn credential to your account")}>
                            <Button
                                variant="outlined"
                                color="primary"
                                onClick={() => {
                                    setShowRegisterDialog(true);
                                }}
                            >
                                {translate("Add")}
                            </Button>
                        </Tooltip>
                    </Grid>
                    <Grid xs={12}>
                        {props.devices === undefined || props.devices.length === 0 ? (
                            <Typography variant={"subtitle2"}>
                                {translate(
                                    "No WebAuthn Credentials have been registered. If you'd like to register one click add.",
                                )}
                            </Typography>
                        ) : (
                            <WebAuthnDevicesStack
                                devices={props.devices}
                                handleRefreshState={props.handleRefreshState}
                            />
                        )}
                    </Grid>
                </Grid>
            </Paper>
        </Fragment>
    );
}
