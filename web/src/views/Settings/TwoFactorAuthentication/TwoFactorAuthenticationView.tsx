import React from "react";

import { Grid } from "@mui/material";

import { AutheliaState } from "@services/State";
import WebauthnDevices from "@views/Settings/TwoFactorAuthentication/WebauthnDevices";

interface Props {
    state: AutheliaState;
}

export default function TwoFactorAuthSettings(props: Props) {
    return (
        <Grid container spacing={2}>
            <Grid item xs={12}>
                <WebauthnDevices state={props.state} />
            </Grid>
        </Grid>
    );
}
