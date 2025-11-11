import { Fragment, ReactNode, useCallback, useEffect, useReducer, useRef } from "react";

import { Box, Button, Link } from "@mui/material";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import FailureIcon from "@components/FailureIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
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

type ComponentState = {
    status: "pushing" | "selecting" | "success" | "failure" | "rate_limited";
    devices: SelectableDevice[];
};

type Action =
    | { type: "set_status"; status: ComponentState["status"] }
    | { type: "set_devices"; devices: SelectableDevice[] }
    | { type: "start_push" }
    | { type: "push_success" }
    | { type: "push_failure" }
    | { type: "select_devices"; devices: SelectableDevice[] }
    | { type: "rate_limited" };

const initialState: ComponentState = {
    status: "pushing",
    devices: [],
};

function reducer(state: ComponentState, action: Action): ComponentState {
    switch (action.type) {
        case "set_status":
            return { ...state, status: action.status };
        case "set_devices":
            return { ...state, devices: action.devices };
        case "start_push":
            return { ...state, status: "pushing" };
        case "push_success":
            return { ...state, status: "success" };
        case "push_failure":
            return { ...state, status: "failure" };
        case "select_devices":
            return { ...state, status: "selecting", devices: action.devices };
        case "rate_limited":
            return { ...state, status: "rate_limited" };
        default:
            return state;
    }
}

export interface Props {
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodMobilePush = function (props: Props) {
    const { t: translate } = useTranslation("portal");
    const { classes } = useStyles();

    const [state, dispatch] = useReducer(reducer, initialState);

    const { createErrorNotification } = useNotifications();

    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        if (timeoutRateLimit.current === null) return;

        return clearTimeout(timeoutRateLimit.current);
    }, []);

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            dispatch({ type: "rate_limited" });

            createErrorNotification(translate("You have made too many requests"));

            timeoutRateLimit.current = setTimeout(() => {
                dispatch({ type: "push_failure" });
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [createErrorNotification, translate],
    );

    const handlePushResponse = useCallback(
        (res: any) => {
            if (res) {
                if (res.data && !res.limited) {
                    switch (res.data.result) {
                        case "auth": {
                            const selectableDevices = [] as SelectableDevice[];
                            for (const d of res.data.devices) {
                                selectableDevices.push({ id: d.device, name: d.display_name, methods: d.capabilities });
                            }
                            dispatch({ type: "select_devices", devices: selectableDevices });
                            break;
                        }
                        case "enroll":
                            createErrorNotification(translate("No compatible device found"));
                            dispatch({ type: "push_failure" });
                            break;
                        case "deny":
                            createErrorNotification(translate("Device selection was denied by Duo policy"));
                            dispatch({ type: "push_failure" });
                            break;
                        default:
                            dispatch({ type: "push_success" });
                            props.onSecondFactorSuccess();
                            break;
                    }
                } else if (res.limited) {
                    handleRateLimited(res.retryAfter);
                } else {
                    createErrorNotification(translate("There was an issue completing sign in process"));
                    dispatch({ type: "push_failure" });
                }
            } else {
                createErrorNotification(translate("There was an issue completing sign in process"));
                dispatch({ type: "push_failure" });
            }
        },
        [createErrorNotification, handleRateLimited, props, translate],
    );

    const handleSelectDevice = useCallback(async () => {
        try {
            const res = await initiateDuoDeviceSelectionProcess();
            switch (res.result) {
                case "auth": {
                    const selectableDevices = [] as SelectableDevice[];
                    for (const d of res.devices) {
                        selectableDevices.push({ id: d.device, name: d.display_name, methods: d.capabilities });
                    }
                    dispatch({ type: "select_devices", devices: selectableDevices });
                    break;
                }
                case "allow":
                    createErrorNotification(translate("Device selection was bypassed by Duo policy"));
                    dispatch({ type: "push_success" });
                    break;
                case "deny":
                    createErrorNotification(translate("Device selection was denied by Duo policy"));
                    dispatch({ type: "push_failure" });
                    break;
                case "enroll":
                    createErrorNotification(translate("No compatible device found"));
                    dispatch({ type: "push_failure" });
                    break;
            }
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("There was an issue fetching Duo device(s)"));
        }
    }, [createErrorNotification, translate]);

    const handleDuoPush = useCallback(async () => {
        try {
            const res = await completePushNotificationSignIn();
            handlePushResponse(res);
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("There was an issue completing sign in process"));
            dispatch({ type: "push_failure" });
        }
    }, [handlePushResponse, createErrorNotification, translate]);

    const updateDuoDevice = useCallback(
        async function (device: DuoDevicePostRequest) {
            try {
                await completeDuoDeviceSelectionProcess(device);
                dispatch({ type: "start_push" });
            } catch (err) {
                console.error(err);
                console.error(new Error(translate("There was an issue updating preferred Duo device")));
            }
        },
        [translate],
    );

    const handleDuoDeviceSelected = useCallback(
        (device: SelectedDevice) => {
            void updateDuoDevice({ device: device.id, method: device.method });
        },
        [updateDuoDevice],
    );

    useEffect(() => {
        if (state.status === "pushing") {
            handleDuoPush();
        }
    }, [state.status, handleDuoPush]);

    if (state.status === "selecting")
        return (
            <DeviceSelectionContainer
                devices={state.devices}
                onBack={() => dispatch({ type: "start_push" })}
                onSelect={handleDuoDeviceSelected}
            />
        );

    let icon: ReactNode;
    switch (state.status) {
        case "pushing":
        case "success":
            icon = <PushNotificationIcon width={64} height={64} animated />;
            break;
        case "failure":
            icon = <FailureIcon />;
    }

    return (
        <Fragment>
            <Box className={classes.container}>
                <Box className={classes.icon}>{icon}</Box>
                <Box className={state.status === "failure" ? "" : "hidden"}>
                    <Button color="secondary" onClick={() => dispatch({ type: "start_push" })}>
                        Retry
                    </Button>
                </Box>
            </Box>
            {state.status === "success" ? null : (
                <Box>
                    <Link component="button" id="selection-link" onClick={handleSelectDevice} underline="hover">
                        {translate("Select a Device")}
                    </Link>
                </Box>
            )}
        </Fragment>
    );
};

const useStyles = makeStyles()(() => ({
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
