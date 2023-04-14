import React from "react";

import { Grid } from "@mui/material";

import { AutheliaState } from "@services/State";
import WebAuthnDevices from "@views/Settings/TwoFactorAuthentication/WebAuthnDevices";

interface Props {
    state: AutheliaState;
}

export default function TwoFactorAuthSettings(props: Props) {
    return (
        <Grid container spacing={2}>
            <Grid item xs={12}>
                <WebAuthnDevices state={props.state} />
            </Grid>
        </Grid>
    );
}
