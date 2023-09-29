import React from "react";

import Grid from "@mui/material/Unstable_Grid2";

import { WebAuthnDevice } from "@models/WebAuthn";
import WebAuthnDeviceItem from "@views/Settings/TwoFactorAuthentication/WebAuthnDeviceItem";

interface Props {
    devices: WebAuthnDevice[];
    handleRefreshState: () => void;
}

export default function WebAuthnDevicesStack(props: Props) {
    return (
        <Grid container spacing={3}>
            {props.devices.map((x, idx) => (
                <WebAuthnDeviceItem key={idx} index={idx} device={x} handleEdit={props.handleRefreshState} />
            ))}
        </Grid>
    );
}
