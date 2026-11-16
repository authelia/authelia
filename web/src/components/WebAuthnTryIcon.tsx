// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useEffect } from "react";

import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";
import FingerTouchIcon from "@components/FingerTouchIcon";
import LinearProgressBar from "@components/LinearProgressBar";
import { Button } from "@components/UI/Button";
import { Spinner } from "@components/UI/Spinner";
import { useTimer } from "@hooks/Timer";
import { WebAuthnCeremony, WebAuthnTouchState } from "@models/WebAuthn";
import IconWithContext from "@views/LoginPortal/SecondFactor/IconWithContext";

interface Props {
    ceremony?: WebAuthnCeremony;
    onMethodSelectionClick?: () => void;
    onRetryClick: () => void;
    webauthnTouchState: WebAuthnTouchState;
}

const WebAuthnTryIcon = function (props: Props) {
    const { t: translate } = useTranslation();
    const timeout = props.ceremony?.timeout ?? 0;
    const [timerPercent, triggerTimer, clearTimer] = useTimer(timeout > 0 ? timeout : 1);

    useEffect(() => {
        if (props.ceremony && timeout > 0) {
            triggerTimer();
        } else {
            clearTimer();
        }
    }, [props.ceremony, timeout, triggerTimer, clearTimer]);

    const touch = (
        <IconWithContext
            icon={<FingerTouchIcon size={64} animated strong />}
            className={props.webauthnTouchState === WebAuthnTouchState.WaitTouch ? undefined : "hidden"}
        >
            <LinearProgressBar value={timerPercent} height={16} />
        </IconWithContext>
    );

    const progress = (
        <IconWithContext
            icon={<Spinner size={48} />}
            className={props.webauthnTouchState === WebAuthnTouchState.InProgress ? undefined : "hidden"}
        >
            <span>{translate("Verifying")}</span>
        </IconWithContext>
    );

    const failure = (
        <IconWithContext
            icon={<FailureIcon />}
            className={
                props.webauthnTouchState === WebAuthnTouchState.Failure ||
                props.webauthnTouchState === WebAuthnTouchState.Unsupported
                    ? undefined
                    : "hidden"
            }
        >
            {props.webauthnTouchState === WebAuthnTouchState.Unsupported ? (
                props.onMethodSelectionClick ? (
                    <Button
                        variant={"ghost"}
                        className="tracking-wide"
                        color={"secondary"}
                        onClick={props.onMethodSelectionClick}
                    >
                        {translate("Use another sign in method")}
                    </Button>
                ) : null
            ) : (
                <Button variant={"ghost"} className="tracking-wide" color={"secondary"} onClick={props.onRetryClick}>
                    {translate("Retry")}
                </Button>
            )}
        </IconWithContext>
    );

    return (
        <div className="inline-block" style={{ minHeight: 101 }}>
            {touch}
            {progress}
            {failure}
        </div>
    );
};

export default WebAuthnTryIcon;
