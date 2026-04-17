import { ReactNode, useCallback, useEffect, useRef, useState } from "react";

import { Box, Button } from "@mui/material";
import axios from "axios";
import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import SuccessIcon from "@components/SuccessIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useAbortSignal } from "@hooks/Abort";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import {
    DuoDevicePostRequest,
    completeDuoDeviceSelectionProcess,
    completePushNotificationSignIn,
    getPreferredDuoDevice,
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

    onSignInError: (_err: Error) => void;
    onSelectionClick: () => void;
    onSignInSuccess: (_redirectURL: string | undefined) => void;
}

const PushNotificationMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();

    const [state, setState] = useState(
        props.authenticationLevel >= AuthenticationLevel.TwoFactor ? State.Success : State.SignInInProgress,
    );
    const redirectionURL = useQueryParam(RedirectionURL);
    const getSignal = useAbortSignal();
    const [enrollUrl, setEnrollUrl] = useState("");
    const [devices, setDevices] = useState([] as SelectableDevice[]);
    const [preferredDevice, setPreferredDevice] = useState<{ device?: string; method?: string }>({});

    const { onSignInError, onSignInSuccess } = props;
    const signInInitiatedRef = useRef(false);
    const stateRef = useRef<null | State>(null);

    const timeoutRateLimit = useRef<NodeJS.Timeout | null>(null);
    const timeoutSuccess = useRef<NodeJS.Timeout | null>(null);

    useEffect(() => {
        return () => {
            if (timeoutRateLimit.current !== null) {
                clearTimeout(timeoutRateLimit.current);
                timeoutRateLimit.current = null;
            }
            if (timeoutSuccess.current !== null) {
                clearTimeout(timeoutSuccess.current);
                timeoutSuccess.current = null;
            }
        };
    }, []);

    useEffect(() => {
        if (props.authenticationLevel < AuthenticationLevel.TwoFactor) {
            getPreferredDuoDevice(getSignal())
                .then((res) => {
                    if (res.preferred_device && res.preferred_method) {
                        setPreferredDevice({ device: res.preferred_device, method: res.preferred_method });
                    }
                })
                .catch((err) => {
                    if (axios.isCancel(err)) return;
                    console.debug("No preferred Duo device found or error fetching:", err);
                });
        }
    }, [props.authenticationLevel, getSignal]);

    const processDevices = useCallback((devices: any[]) => {
        return devices.map((d: { device: any; display_name: any; capabilities: any }) => ({
            id: d.device,
            methods: d.capabilities,
            name: d.display_name,
        }));
    }, []);

    const handleSuccess = useCallback(
        (redirect: string | undefined) => {
            setState(State.Success);
            timeoutSuccess.current = setTimeout(() => {
                onSignInSuccess(redirect);
                timeoutSuccess.current = null;
            }, 1500);
        },
        [onSignInSuccess],
    );

    const handleRateLimited = useCallback(
        (retryAfter: number) => {
            if (timeoutRateLimit.current) {
                clearTimeout(timeoutRateLimit.current);
            }

            setState(State.RateLimited);

            onSignInError(new Error(translate("You have made too many requests")));

            timeoutRateLimit.current = setTimeout(() => {
                setState(State.Failure);
                timeoutRateLimit.current = null;
            }, retryAfter * 1000);
        },
        [onSignInError, translate],
    );

    const handleFetchDuoDevices = useCallback(async () => {
        try {
            const res = await initiateDuoDeviceSelectionProcess(getSignal());
            if (res.preferred_device && res.preferred_method) {
                setPreferredDevice({ device: res.preferred_device, method: res.preferred_method });
            }
            switch (res.result) {
                case "auth": {
                    setDevices(processDevices(res.devices));
                    stateRef.current = state;
                    setState(State.Selection);
                    break;
                }
                case "allow":
                    onSignInError(new Error(translate("Device selection was bypassed by Duo policy")));
                    setState(State.Success);
                    break;
                case "deny":
                    onSignInError(new Error(translate("Device selection was denied by Duo policy")));
                    setState(State.Failure);
                    break;
                case "enroll":
                    onSignInError(new Error(translate("No compatible device found")));
                    if (res.enroll_url && props.duoSelfEnrollment) setEnrollUrl(res.enroll_url);
                    setState(State.Enroll);
                    break;
            }
        } catch (err) {
            if (axios.isCancel(err)) return;
            console.error(err);
            onSignInError(new Error(translate("There was an issue fetching Duo device(s)")));
        }
    }, [getSignal, onSignInError, translate, props.duoSelfEnrollment, processDevices, state]);

    const handleSignIn = useCallback(async () => {
        if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        try {
            setState(State.SignInInProgress);
            const res = await completePushNotificationSignIn(
                redirectionURL,
                flowID,
                flow,
                subflow,
                userCode,
                getSignal(),
            );
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
                    onSignInError(new Error(translate("No compatible device found")));
                    if (res.data.enroll_url && props.duoSelfEnrollment) setEnrollUrl(res.data.enroll_url);
                    setState(State.Enroll);
                    break;
                case "deny":
                    onSignInError(new Error(translate("Device selection was denied by Duo policy")));
                    setState(State.Failure);
                    break;
                default:
                    handleSuccess(res.data.redirect);
            }
        } catch (err) {
            if (axios.isCancel(err)) return;
            console.error(err);
            onSignInError(new Error(translate("There was an issue completing sign in process")));
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
        getSignal,
        onSignInError,
        translate,
        handleRateLimited,
        processDevices,
        handleSuccess,
    ]);

    const updateDuoDevice = useCallback(
        async function (device: DuoDevicePostRequest) {
            try {
                await completeDuoDeviceSelectionProcess(device, getSignal());
                if (!props.registered) {
                    props.onSelectionClick();
                }
            } catch (err) {
                if (axios.isCancel(err)) return;
                console.error(err);
                onSignInError(new Error(translate("There was an issue updating preferred Duo device")));
            }
        },
        [getSignal, onSignInError, props, translate],
    );

    const handleDuoDeviceSelected = useCallback(
        (device: SelectedDevice) => {
            const selected = { device: device.id, method: device.method };
            const isDifferent =
                selected.device !== preferredDevice?.device || selected.method !== preferredDevice?.method;

            setPreferredDevice(selected);

            if (isDifferent) {
                updateDuoDevice(selected)
                    .then(() => {
                        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
                            setState(State.Success);
                        } else {
                            handleSignIn();
                        }
                    })
                    .catch(() => {
                        setState(State.Failure);
                    });
            } else if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
                setState(State.Success);
            } else {
                setState(State.SignInInProgress);
            }
        },
        [updateDuoDevice, preferredDevice, handleSignIn, setPreferredDevice, props.authenticationLevel],
    );

    useEffect(() => {
        if (props.authenticationLevel < AuthenticationLevel.TwoFactor && !signInInitiatedRef.current) {
            signInInitiatedRef.current = true;
            handleSignIn();
        }
    }, [props.authenticationLevel, handleSignIn]);

    if (state === State.Selection)
        return (
            <DeviceSelectionContainer
                devices={devices}
                onBack={() => setState(stateRef.current || State.SignInInProgress)}
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
            onRegisterClick={() => window.open(enrollUrl, "_blank", "noopener,noreferrer")}
        >
            <Box sx={{ display: "inline-block", height: "64px", width: "64px" }}>{icon}</Box>
            <Box className={state === State.Failure ? "" : "hidden"}>
                <Button color="secondary" onClick={handleSignIn}>
                    {translate("Retry")}
                </Button>
            </Box>
        </MethodContainer>
    );
};

export default PushNotificationMethod;
