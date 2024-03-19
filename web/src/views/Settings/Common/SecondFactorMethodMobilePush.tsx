import React, { Fragment, ReactNode, useCallback, useEffect, useState } from "react";

import { Box, Button, Link, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import { useIsMountedRef } from "@hooks/Mounted";
import {
    DuoDevicePostRequest,
    completeDuoDeviceSelectionProcess,
    completePushNotificationSignIn,
    initiateDuoDeviceSelectionProcess,
} from "@services/PushNotification";
import DeviceSelectionContainer, {
    SelectableDevice,
    SelectedDevice,
} from "@views/LoginPortal/SecondFactor/DeviceSelectionContainer";

export enum State {
    SignInInProgress = 1,
    Success = 2,
    Failure = 3,
    Selection = 4,
}

export interface Props {
    closing: boolean;
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodMobilePush = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    const styles = useStyles();
    const [state, setState] = useState(State.SignInInProgress);
    const mounted = useIsMountedRef();
    const [devices, setDevices] = useState([] as SelectableDevice[]);

    const handleSelectDevice = useCallback(async () => {
        try {
            const res = await initiateDuoDeviceSelectionProcess();
            if (!mounted.current) return;
            switch (res.result) {
                case "auth":
                    let selectableDevices = [] as SelectableDevice[];
                    res.devices.forEach((d: { device: any; display_name: any; capabilities: any }) =>
                        selectableDevices.push({ id: d.device, name: d.display_name, methods: d.capabilities }),
                    );
                    setDevices(selectableDevices);
                    setState(State.Selection);
                    break;
                case "allow":
                    console.error(new Error(translate("Device selection was bypassed by Duo policy")));
                    setState(State.Success);
                    break;
                case "deny":
                    console.error(new Error(translate("Device selection was denied by Duo policy")));
                    setState(State.Failure);
                    break;
                case "enroll":
                    console.error(new Error(translate("No compatible device found")));
                    setState(State.Failure);
                    break;
            }
        } catch (err) {
            if (!mounted.current) return;
            console.error(err);
            console.error(new Error(translate("There was an issue fetching Duo device(s)")));
        }
    }, [mounted, translate]);

    const handleDuoPush = useCallback(async () => {
        try {
            setState(State.SignInInProgress);
            const res = await completePushNotificationSignIn();
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            if (res) {
                switch (res.result) {
                    case "auth":
                        let selectableDevices = [] as SelectableDevice[];
                        res.devices.forEach((d) =>
                            selectableDevices.push({ id: d.device, name: d.display_name, methods: d.capabilities }),
                        );
                        setDevices(selectableDevices);
                        setState(State.Selection);
                        return;
                    case "enroll":
                        console.error(new Error(translate("No compatible device found")));
                        setState(State.Failure);
                        return;
                    case "deny":
                        console.error(new Error(translate("Device selection was denied by Duo policy")));
                        setState(State.Failure);
                        return;
                }
            }

            setState(State.Success);
            props.onSecondFactorSuccess();
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current || state !== State.SignInInProgress) return;

            console.error(err);
            console.error(new Error(translate("There was an issue completing sign in process")));
            setState(State.Failure);
        }
    }, [mounted, props, state, translate]);

    const updateDuoDevice = useCallback(
        async function (device: DuoDevicePostRequest) {
            try {
                await completeDuoDeviceSelectionProcess(device);
                setState(State.SignInInProgress);
            } catch (err) {
                console.error(err);
                console.error(new Error(translate("There was an issue updating preferred Duo device")));
            }
        },
        [translate],
    );

    const handleDuoDeviceSelected = useCallback(
        (device: SelectedDevice) => {
            updateDuoDevice({ device: device.id, method: device.method });
        },
        [updateDuoDevice],
    );

    useEffect(() => {
        if (state === State.SignInInProgress) handleDuoPush();
    }, [handleDuoPush, state]);

    if (state === State.Selection)
        return (
            <DeviceSelectionContainer
                devices={devices}
                onBack={() => setState(State.SignInInProgress)}
                onSelect={handleDuoDeviceSelected}
            />
        );

    let icon: ReactNode;
    switch (state) {
        case State.SignInInProgress:
        case State.Success:
            icon = <PushNotificationIcon width={64} height={64} animated />;
            break;
        case State.Failure:
            icon = <FailureIcon />;
    }

    return (
        <Fragment>
            <Box className={styles.container}>
                <Box className={styles.icon}>{icon}</Box>
                <Box className={state !== State.Failure ? "hidden" : ""}>
                    <Button color="secondary" onClick={handleDuoPush}>
                        Retry
                    </Button>
                </Box>
            </Box>
            {state !== State.Success ? (
                <Box>
                    <Link component="button" id="selection-link" onClick={handleSelectDevice} underline="hover">
                        {translate("Select a Device")}
                    </Link>
                </Box>
            ) : null}
        </Fragment>
    );
};

export default SecondFactorMethodMobilePush;

const useStyles = makeStyles((theme: Theme) => ({
    container: {
        height: "120px",
    },
    icon: {
        width: "64px",
        height: "64px",
        display: "inline-block",
    },
}));
