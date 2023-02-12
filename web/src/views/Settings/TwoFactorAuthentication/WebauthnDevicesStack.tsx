import React, { Fragment, useEffect, useState } from "react";

import { Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import { WebauthnDevice } from "@models/Webauthn";
import { getWebauthnDevices } from "@services/UserWebauthnDevices";
import WebauthnDeviceItem from "@views/Settings/TwoFactorAuthentication/WebauthnDeviceItem";

interface Props {
    refreshState: number;
    incrementRefreshState: () => void;
}

export default function WebauthnDevicesStack(props: Props) {
    const { t: translate } = useTranslation("settings");

    const [devices, setDevices] = useState<WebauthnDevice[] | null>(null);

    useEffect(() => {
        (async function () {
            setDevices(null);
            const devices = await getWebauthnDevices();
            setDevices(devices);
        })();
    }, [props.refreshState]);

    return (
        <Fragment>
            {devices !== null && devices.length !== 0 ? (
                <Stack spacing={3}>
                    {devices.map((x, idx) => (
                        <WebauthnDeviceItem key={idx} index={idx} device={x} handleEdit={props.incrementRefreshState} />
                    ))}
                </Stack>
            ) : (
                <Typography>{translate("No Registered Webauthn Credentials")}</Typography>
            )}
        </Fragment>
    );
}
