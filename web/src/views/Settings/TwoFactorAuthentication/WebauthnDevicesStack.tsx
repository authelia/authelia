import React, { useEffect, useState } from "react";

import { Stack } from "@mui/material";

import { WebauthnDevice } from "@models/Webauthn";
import { getWebauthnDevices } from "@services/UserWebauthnDevices";
import WebauthnDeviceItem from "@views/Settings/TwoFactorAuthentication/WebauthnDeviceItem";

interface Props {}

export default function WebauthnDevicesStack(props: Props) {
    const [devices, setDevices] = useState<WebauthnDevice[]>([]);

    useEffect(() => {
        (async function () {
            const devices = await getWebauthnDevices();
            setDevices(devices);
        })();
    }, []);

    const handleEdit = (index: number, device: WebauthnDevice) => {
        const nextDevices = devices.map((d, i) => {
            if (i === index) {
                return device;
            } else {
                return d;
            }
        });

        setDevices(nextDevices);
    };

    const handleDelete = (device: WebauthnDevice) => {
        setDevices(devices.filter((d) => d.id !== device.id && d.kid !== device.kid));
    };

    return (
        <Stack spacing={3}>
            {devices
                ? devices.map((x, idx) => (
                      <WebauthnDeviceItem
                          index={idx}
                          device={x}
                          handleDeviceEdit={handleEdit}
                          handleDeviceDelete={handleDelete}
                      />
                  ))
                : null}
        </Stack>
    );
}
