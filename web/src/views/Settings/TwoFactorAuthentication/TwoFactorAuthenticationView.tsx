import React from "react";

import { Grid } from "@mui/material";

import SettingsLayout from "@layouts/SettingsLayout";
import { AutheliaState } from "@services/State";

import TOTP from "./TOTP";
import WebauthnDevices from "./WebauthnDevices";

interface Props {
    state: AutheliaState;
}

export default function TwoFactorAuthSettings(props: Props) {
    return (
        <SettingsLayout>
            <Grid container spacing={2}>
                <Grid item xs={12}>
                    <WebauthnDevices state={props.state} />
                </Grid>
                <Grid item xs={12}>
                    <TOTP state={props.state} />
                </Grid>
            </Grid>
        </SettingsLayout>
    );
}
