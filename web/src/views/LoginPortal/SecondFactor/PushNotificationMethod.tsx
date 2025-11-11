import { ReactNode, useCallback, useEffect, useRef, useState } from "react";

import { Box, Button } from "@mui/material";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import FailureIcon from "@components/FailureIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import SuccessIcon from "@components/SuccessIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useIsMountedRef } from "@hooks/Mounted";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import {
    DuoDevicePostRequest,
    completeDuoDeviceSelectionProcess,
    completePushNotificationSignIn,
    initiateDuoDeviceSelectionProcess,
} from "@services/PushNotification";
import { AuthenticationLevel } from "@services/State";
import DeviceSelectionContainer, {
    SelectableDevice,
    SelectedDevice,
} from "@views/LoginPortal/SecondFactor/DeviceSelectionContainer";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";

export enum State {
    SignInInProgress = 1,
    Success = 2,
    Failure = 3,
    Selection = 4,
    Enroll = 5,
    RateLimited = 6,
}

export interface Props {
    id: string;
    authenticationLevel: AuthenticationLevel;
    duoSelfEnrollment: boolean;
    registered: boolean;

    onSignInError: (err: Error) => void;
    onSelectionClick: () => void;
    onSignInSuccess: (redirectURL: string | undefined) => void;
}

const PushNotificationMethod = function (props: Props) {
    const { t: translate } = useTranslation();
    const { classes } = useStyles();

    const { id: flowID, flow, subflow } = useFlow();
    const userCode = useUserCode();

    const [state, setState] = useState(
        props.authenticationLevel >= AuthenticationLevel.TwoFactor ? State.Success : State.SignInInProgress,
    );
    const redirectionURL = useQueryParam(RedirectionURL);
    const mounted = useIsMountedRef();
    const [enrollUrl, setEnrollUrl] = useState("");
    const [devices, setDevices] = useState([] as SelectableDevice[]);

    const { onSignInSuccess, onSignInError } = props;
    const onSignInErrorCallback = useRef(onSignInError).current;
    const onSignInSuccessCallback = useRef(onSignInSuccess).current;

    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        if (timeoutRateLimit.current === null) return;

        return clearTimeout(timeoutRateLimit.current);
    }, []);

    const processDevices = useCallback((devices: any[]) => {
        return devices.map((d: { device: any; display_name: any; capabilities: any }) => ({
            id: d.device,
            name: d.display_name,
            methods: d.capabilities,
        }));
    }, []);

    const handleSuccess = useCallback(
        (redirect: string | undefined) => {
            setState(State.Success);
            setTimeout(() => {
                if (!mounted.current) return;
                onSignInSuccessCallback(redirect);
            }, 1500);
        },
        [mounted, onSignInSuccessCallback],
    );

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            setState(State.RateLimited);

            onSignInErrorCallback(new Error(translate("You have made too many requests")));

            timeoutRateLimit.current = setTimeout(() => {
                setState(State.Failure);
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [onSignInErrorCallback, translate],
    );

    const handleFetchDuoDevices = async () => {
        try {
            const res = await initiateDuoDeviceSelectionProcess();
            if (!mounted.current) return;
            switch (res.result) {
                case "auth": {
                    let selectableDevices = [] as SelectableDevice[];
                    for (const d of res.devices) {
                        selectableDevices.push({ id: d.device, name: d.display_name, methods: d.capabilities });
                    }
                    setDevices(selectableDevices);
                    setState(State.Selection);
                    break;
                }
                case "allow":
                    onSignInErrorCallback(new Error(translate("Device selection was bypassed by Duo policy")));
                    setState(State.Success);
                    break;
                case "deny":
                    onSignInErrorCallback(new Error(translate("Device selection was denied by Duo policy")));
                    setState(State.Failure);
                    break;
                case "enroll":
                    onSignInErrorCallback(new Error(translate("No compatible device found")));
                    if (res.enroll_url && props.duoSelfEnrollment) setEnrollUrl(res.enroll_url);
                    setState(State.Enroll);
                    break;
            }
        } catch (err) {
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error(translate("There was an issue fetching Duo device(s)")));
        }
    };

    const handleSignIn = useCallback(async () => {
        if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        try {
            setState(State.SignInInProgress);
            const res = await completePushNotificationSignIn(redirectionURL, flowID, flow, subflow, userCode);
            if (!mounted.current) return;
            if (!res) {
                throw new Error(translate("There was an issue completing sign in process"));
            }
            if (res.limited) {
                handleRateLimited(res.retryAfter);
                return;
            }
            if (!res.data) {
                handleSuccess(undefined);
                return;
            }
            switch (res.data.result) {
                case "auth":
                    setDevices(processDevices(res.data.devices));
                    setState(State.Selection);
                    break;
                case "enroll":
                    onSignInErrorCallback(new Error(translate("No compatible device found")));
                    if (res.data.enroll_url && props.duoSelfEnrollment) setEnrollUrl(res.data.enroll_url);
                    setState(State.Enroll);
                    break;
                case "deny":
                    onSignInErrorCallback(new Error(translate("Device selection was denied by Duo policy")));
                    setState(State.Failure);
                    break;
                default:
                    handleSuccess(res.data.redirect);
            }
        } catch (err) {
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error(translate("There was an issue completing sign in process")));
            setState(State.Failure);
        }
    }, [
        props.authenticationLevel,
        props.duoSelfEnrollment,
        redirectionURL,
        flowID,
        flow,
        subflow,
        userCode,
        mounted,
        onSignInErrorCallback,
        translate,
        handleRateLimited,
        processDevices,
        handleSuccess,
    ]);

    const updateDuoDevice = useCallback(
        async function (device: DuoDevicePostRequest) {
            try {
                await completeDuoDeviceSelectionProcess(device);
                if (props.registered) {
                    setState(State.SignInInProgress);
                    await handleSignIn();
                } else {
                    setState(State.Failure);
                    props.onSelectionClick();
                }
            } catch (err) {
                console.error(err);
                onSignInErrorCallback(new Error(translate("There was an issue updating preferred Duo device")));
            }
        },
        [onSignInErrorCallback, props, translate, handleSignIn],
    );

    const handleDuoDeviceSelected = useCallback(
        (device: SelectedDevice) => {
            updateDuoDevice({ device: device.id, method: device.method });
        },
        [updateDuoDevice],
    );

    useEffect(() => {
        if (props.authenticationLevel < AuthenticationLevel.TwoFactor) {
            handleSignIn();
        }
    }, [props.authenticationLevel, handleSignIn]);

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
            icon = <PushNotificationIcon width={64} height={64} animated />;
            break;
        case State.Success:
            icon = <SuccessIcon />;
            break;
        case State.Failure:
            icon = <FailureIcon />;
    }

    let methodState = MethodContainerState.METHOD;
    if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
        methodState = MethodContainerState.ALREADY_AUTHENTICATED;
    } else if (state === State.Enroll) {
        methodState = MethodContainerState.NOT_REGISTERED;
    }

    return (
        <MethodContainer
            id={props.id}
            title={translate("Push Notification")}
            explanation={translate("A notification has been sent to your smartphone")}
            duoSelfEnrollment={enrollUrl ? props.duoSelfEnrollment : false}
            registered={props.registered}
            state={methodState}
            onSelectClick={handleFetchDuoDevices}
            onRegisterClick={() => window.open(enrollUrl, "_blank")}
        >
            <Box className={classes.icon}>{icon}</Box>
            <Box className={state === State.Failure ? "" : "hidden"}>
                <Button color="secondary" onClick={handleSignIn}>
                    {translate("Retry")}
                </Button>
            </Box>
        </MethodContainer>
    );
};

const useStyles = makeStyles()(() => ({
    icon: {
        width: "64px",
        height: "64px",
        display: "inline-block",
    },
}));

export default PushNotificationMethod;
