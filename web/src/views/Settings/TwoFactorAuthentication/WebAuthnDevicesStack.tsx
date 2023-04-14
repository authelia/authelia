import React, { Fragment, useEffect, useState } from "react";

import { Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import { WebAuthnDevice } from "@models/WebAuthn";
import { getWebAuthnDevices } from "@services/UserWebAuthnDevices";
import WebAuthnDeviceItem from "@views/Settings/TwoFactorAuthentication/WebAuthnDeviceItem";

interface Props {
    refreshState: number;
    incrementRefreshState: () => void;
}

export default function WebAuthnDevicesStack(props: Props) {
    const { t: translate } = useTranslation("settings");

    const [devices, setDevices] = useState<WebAuthnDevice[] | null>(null);

    useEffect(() => {
        (async function () {
            setDevices(null);
            const devices = await getWebAuthnDevices();
            setDevices(devices);
        })();
    }, [props.refreshState]);

    return (
        <Fragment>
            {devices !== null && devices.length !== 0 ? (
                <Stack spacing={3}>
                    {devices.map((x, idx) => (
                        <WebAuthnDeviceItem key={idx} index={idx} device={x} handleEdit={props.incrementRefreshState} />
                    ))}
                </Stack>
            ) : (
                <Typography>{translate("No Registered WebAuthn Credentials")}</Typography>
            )}
        </Fragment>
    );
}
