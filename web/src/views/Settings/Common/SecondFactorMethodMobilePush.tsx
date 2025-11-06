import React, { Fragment, ReactNode, useCallback, useEffect, useRef, useState } from "react";

import { Box, Button, Link, Theme } from "@mui/material";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import FailureIcon from "@components/FailureIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import { useIsMountedRef } from "@hooks/Mounted";
import { useNotifications } from "@hooks/NotificationsContext";
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
    RateLimited = 5,
}

export interface Props {
    closing: boolean;
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodMobilePush = function (props: Props) {
    const { t: translate } = useTranslation("portal");
    const { classes } = useStyles();

    const [state, setState] = useState(State.SignInInProgress);
    const mounted = useIsMountedRef();
    const [devices, setDevices] = useState([] as SelectableDevice[]);

    const { createErrorNotification } = useNotifications();

    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        return () => {
            if (timeoutRateLimit.current !== null) {
                clearTimeout(timeoutRateLimit.current);
                timeoutRateLimit.current = null;
            }
        };
    }, []);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (!mounted.current) {
                return;
            }

            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            setState(State.RateLimited);

            createErrorNotification(translate("You have made too many requests"));

            timeoutRateLimit.current = setTimeout(() => {
                if (!mounted.current) {
                    timeoutRateLimit.current = null;
                    return;
                }
                setState(State.Failure);
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [createErrorNotification, mounted, translate],
    );

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
                    createErrorNotification(translate("Device selection was bypassed by Duo policy"));
                    setState(State.Success);
                    break;
                case "deny":
                    createErrorNotification(translate("Device selection was denied by Duo policy"));
                    setState(State.Failure);
                    break;
                case "enroll":
                    createErrorNotification(translate("No compatible device found"));
                    setState(State.Failure);
                    break;
            }
        } catch (err) {
            if (!mounted.current) return;
            console.error(err);
            createErrorNotification(translate("There was an issue fetching Duo device(s)"));
        }
    }, [createErrorNotification, mounted, translate]);

    const handleDuoPush = useCallback(async () => {
        try {
            setState(State.SignInInProgress);
            const res = await completePushNotificationSignIn();
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;

            if (res) {
                if (res.data && !res.limited) {
                    switch (res.data.result) {
                        case "auth":
                            let selectableDevices = [] as SelectableDevice[];
                            res.data.devices.forEach((d) =>
                                selectableDevices.push({ id: d.device, name: d.display_name, methods: d.capabilities }),
                            );
                            setDevices(selectableDevices);
                            setState(State.Selection);
                            break;
                        case "enroll":
                            createErrorNotification(translate("No compatible device found"));
                            setState(State.Failure);
                            break;
                        case "deny":
                            createErrorNotification(translate("Device selection was denied by Duo policy"));
                            setState(State.Failure);
                            break;
                        default:
                            setState(State.Success);
                            props.onSecondFactorSuccess();
                            break;
                    }
                } else if (res.limited) {
                    handleRateLimited(res.retryAfter);
                } else {
                    createErrorNotification(translate("There was an issue completing sign in process"));
                    setState(State.Failure);
                }
            } else {
                createErrorNotification(translate("There was an issue completing sign in process"));
                setState(State.Failure);
            }
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current || state !== State.SignInInProgress) return;

            console.error(err);
            createErrorNotification(translate("There was an issue completing sign in process"));
            setState(State.Failure);
        }
    }, [createErrorNotification, handleRateLimited, mounted, props, state, translate]);

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
            <Box className={classes.container}>
                <Box className={classes.icon}>{icon}</Box>
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

const useStyles = makeStyles()((theme: Theme) => ({
    container: {
        height: "120px",
    },
    icon: {
        width: "64px",
        height: "64px",
        display: "inline-block",
    },
}));

export default SecondFactorMethodMobilePush;
