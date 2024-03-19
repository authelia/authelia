import React, { ReactNode, useCallback, useEffect, useRef, useState } from "react";

import { Button, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import SuccessIcon from "@components/SuccessIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useIsMountedRef } from "@hooks/Mounted";
import { useQueryParam } from "@hooks/QueryParam";
import { useWorkflow } from "@hooks/Workflow";
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
    const styles = useStyles();
    const [state, setState] = useState(State.SignInInProgress);
    const redirectionURL = useQueryParam(RedirectionURL);
    const [workflow, workflowID] = useWorkflow();
    const mounted = useIsMountedRef();
    const [enroll_url, setEnrollUrl] = useState("");
    const [devices, setDevices] = useState([] as SelectableDevice[]);
    const { t: translate } = useTranslation();

    const { onSignInSuccess, onSignInError } = props;
    const onSignInErrorCallback = useRef(onSignInError).current;
    const onSignInSuccessCallback = useRef(onSignInSuccess).current;

    const fetchDuoDevicesFunc = useCallback(async () => {
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
    }, [props.duoSelfEnrollment, mounted, onSignInErrorCallback, translate]);

    const signInFunc = useCallback(async () => {
        if (props.authenticationLevel === AuthenticationLevel.TwoFactor) {
            return;
        }

        try {
            setState(State.SignInInProgress);
            const res = await completePushNotificationSignIn(redirectionURL, workflow, workflowID);
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            if (res && res.result === "auth") {
                let selectableDevices = [] as SelectableDevice[];
                res.devices.forEach((d) =>
                    selectableDevices.push({ id: d.device, name: d.display_name, methods: d.capabilities }),
                );
                setDevices(selectableDevices);
                setState(State.Selection);
                return;
            }
            if (res && res.result === "enroll") {
                onSignInErrorCallback(new Error(translate("No compatible device found")));
                if (res.enroll_url && props.duoSelfEnrollment) setEnrollUrl(res.enroll_url);
                setState(State.Enroll);
                return;
            }
            if (res && res.result === "deny") {
                onSignInErrorCallback(new Error(translate("Device selection was denied by Duo policy")));
                setState(State.Failure);
                return;
            }

            setState(State.Success);
            setTimeout(() => {
                if (!mounted.current) return;
                onSignInSuccessCallback(res ? res.redirect : undefined);
            }, 1500);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current || state !== State.SignInInProgress) return;

            console.error(err);
            onSignInErrorCallback(new Error(translate("There was an issue completing sign in process")));
            setState(State.Failure);
        }
    }, [
        props.authenticationLevel,
        props.duoSelfEnrollment,
        redirectionURL,
        workflow,
        workflowID,
        mounted,
        onSignInErrorCallback,
        onSignInSuccessCallback,
        state,
        translate,
    ]);

    const updateDuoDevice = useCallback(
        async function (device: DuoDevicePostRequest) {
            try {
                await completeDuoDeviceSelectionProcess(device);
                if (!props.registered) {
                    setState(State.SignInInProgress);
                    props.onSelectionClick();
                } else {
                    setState(State.SignInInProgress);
                }
            } catch (err) {
                console.error(err);
                onSignInErrorCallback(new Error(translate("There was an issue updating preferred Duo device")));
            }
        },
        [onSignInErrorCallback, props, translate],
    );

    const handleDuoDeviceSelected = useCallback(
        (device: SelectedDevice) => {
            updateDuoDevice({ device: device.id, method: device.method });
        },
        [updateDuoDevice],
    );

    // Set successful state if user is already authenticated.
    useEffect(() => {
        if (props.authenticationLevel >= AuthenticationLevel.TwoFactor) {
            setState(State.Success);
        }
    }, [props.authenticationLevel, setState]);

    useEffect(() => {
        if (state === State.SignInInProgress) signInFunc();
    }, [signInFunc, state]);

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
            title="Push Notification"
            explanation="A notification has been sent to your smartphone"
            duoSelfEnrollment={enroll_url ? props.duoSelfEnrollment : false}
            registered={props.registered}
            state={methodState}
            onSelectClick={fetchDuoDevicesFunc}
            onRegisterClick={() => window.open(enroll_url, "_blank")}
        >
            <div className={styles.icon}>{icon}</div>
            <div className={state !== State.Failure ? "hidden" : ""}>
                <Button color="secondary" onClick={signInFunc}>
                    Retry
                </Button>
            </div>
        </MethodContainer>
    );
};

export default PushNotificationMethod;

const useStyles = makeStyles((theme: Theme) => ({
    icon: {
        width: "64px",
        height: "64px",
        display: "inline-block",
    },
}));
