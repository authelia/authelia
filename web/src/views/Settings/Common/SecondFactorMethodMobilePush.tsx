import { Fragment, ReactNode, useCallback, useEffect, useReducer, useRef } from "react";

import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import { Button } from "@components/UI/Button";
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
    status: "failure" | "pushing" | "rate_limited" | "selecting" | "success";
    devices: SelectableDevice[];
};

type Action =
    | { type: "pushFailure" }
    | { type: "pushSuccess" }
    | { type: "rateLimited" }
    | { type: "selectDevices"; devices: SelectableDevice[] }
    | { type: "setDevices"; devices: SelectableDevice[] }
    | { type: "setStatus"; status: ComponentState["status"] }
    | { type: "startPush" };

const initialState: ComponentState = {
    devices: [],
    status: "pushing",
};

function reducer(state: ComponentState, action: Action): ComponentState {
    switch (action.type) {
        case "setStatus":
            return { ...state, status: action.status };
        case "setDevices":
            return { ...state, devices: action.devices };
        case "startPush":
            return { ...state, status: "pushing" };
        case "pushSuccess":
            return { ...state, status: "success" };
        case "pushFailure":
            return { ...state, status: "failure" };
        case "selectDevices":
            return { ...state, devices: action.devices, status: "selecting" };
        case "rateLimited":
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

    const [state, dispatch] = useReducer(reducer, initialState);

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
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            dispatch({ type: "rateLimited" });

            createErrorNotification(translate("You have made too many requests"));

            timeoutRateLimit.current = setTimeout(() => {
                dispatch({ type: "pushFailure" });
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
                                selectableDevices.push({ id: d.device, methods: d.capabilities, name: d.display_name });
                            }
                            dispatch({ devices: selectableDevices, type: "selectDevices" });
                            break;
                        }
                        case "enroll":
                            createErrorNotification(translate("No compatible device found"));
                            dispatch({ type: "pushFailure" });
                            break;
                        case "deny":
                            createErrorNotification(translate("Device selection was denied by Duo policy"));
                            dispatch({ type: "pushFailure" });
                            break;
                        default:
                            dispatch({ type: "pushSuccess" });
                            props.onSecondFactorSuccess();
                            break;
                    }
                } else if (res.limited) {
                    handleRateLimited(res.retryAfter);
                } else {
                    createErrorNotification(translate("There was an issue completing sign in process"));
                    dispatch({ type: "pushFailure" });
                }
            } else {
                createErrorNotification(translate("There was an issue completing sign in process"));
                dispatch({ type: "pushFailure" });
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
                        selectableDevices.push({ id: d.device, methods: d.capabilities, name: d.display_name });
                    }
                    dispatch({ devices: selectableDevices, type: "selectDevices" });
                    break;
                }
                case "allow":
                    createErrorNotification(translate("Device selection was bypassed by Duo policy"));
                    dispatch({ type: "pushSuccess" });
                    break;
                case "deny":
                    createErrorNotification(translate("Device selection was denied by Duo policy"));
                    dispatch({ type: "pushFailure" });
                    break;
                case "enroll":
                    createErrorNotification(translate("No compatible device found"));
                    dispatch({ type: "pushFailure" });
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
            dispatch({ type: "pushFailure" });
        }
    }, [handlePushResponse, createErrorNotification, translate]);

    const updateDuoDevice = useCallback(
        async function (device: DuoDevicePostRequest) {
            try {
                await completeDuoDeviceSelectionProcess(device);
                dispatch({ type: "startPush" });
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
        if (state.status === "pushing") {
            handleDuoPush();
        }
    }, [state.status, handleDuoPush]);

    if (state.status === "selecting")
        return (
            <DeviceSelectionContainer
                devices={state.devices}
                onBack={() => dispatch({ type: "startPush" })}
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
            <div className="h-[120px]">
                <div className="inline-block h-16 w-16">{icon}</div>
                <div className={state.status === "failure" ? "" : "hidden"}>
                    <Button variant="secondary" onClick={() => dispatch({ type: "startPush" })}>
                        Retry
                    </Button>
                </div>
            </div>
            {state.status === "success" ? null : (
                <div>
                    <button
                        className="text-primary underline-offset-4 hover:underline bg-transparent border-none cursor-pointer text-sm"
                        id="selection-link"
                        onClick={handleSelectDevice}
                    >
                        {translate("Select a Device")}
                    </button>
                </div>
            )}
        </Fragment>
    );
};

export default SecondFactorMethodMobilePush;
