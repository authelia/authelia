import React, { useEffect, useState } from "react";

import { Grid } from "@mui/material";

import { WebauthnDevice } from "@root/models/Webauthn";
import { getWebauthnDevices } from "@root/services/UserWebauthnDevices";
import { AutheliaState } from "@services/State";

import WebauthnDevices from "./WebauthnDevices";

interface Props {
    state: AutheliaState;
}

export default function TwoFactorAuthSettings(props: Props) {
    const [webauthnDevices, setWebauthnDevices] = useState<WebauthnDevice[] | undefined>();

    useEffect(() => {
        (async function () {
            const devices = await getWebauthnDevices();
            setWebauthnDevices(devices);
        })();
    }, []);

    return (
        <Grid container spacing={2}>
            <Grid item xs={12}>
                <WebauthnDevices state={props.state} webauthnDevices={webauthnDevices} />
            </Grid>
        </Grid>
    );
}
