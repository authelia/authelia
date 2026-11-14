// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useEffectEvent, useState } from "react";

import axios from "axios";
import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";
import RateLimitCountdown from "@components/RateLimitCountdown";
import SuccessIcon from "@components/SuccessIcon";
import { Button } from "@components/UI/Button";
import { Spinner } from "@components/UI/Spinner";
import { useAbortSignal } from "@hooks/Abort";
import { useSecondsUntil } from "@hooks/Countdown";
import { useRateLimit } from "@hooks/RateLimit";
import {
    AppAuthenticatorSignInBody,
    AppAuthenticatorSignInStatus,
    AppAuthenticatorChallenge as Challenge,
    getAppAuthenticatorSignInStatus,
    initiateAppAuthenticatorSignIn,
} from "@services/AppAuthenticator";
import { RateLimitedError } from "@services/RateLimit";

const PollIntervalMs = 2000;

type Status = "denied" | "failure" | "not_approved" | "pending" | "rate_limited" | "starting" | "success";

export interface Props {
    body: AppAuthenticatorSignInBody;

    onSuccess: (_redirectURL: string | undefined) => void;
    onError: (_err: Error) => void;
}

const AppAuthenticatorChallenge = function (props: Props) {
    const { t: translate } = useTranslation("portal");

    const getSignal = useAbortSignal();
    const [status, setStatus] = useState<Status>("starting");
    const [challenge, setChallenge] = useState<Challenge>();
    const [poll, setPoll] = useState(0);

    const seconds = useSecondsUntil(status === "pending" ? challenge?.expires_at : undefined);
    const { secondsRemaining, trigger: triggerRateLimit } = useRateLimit(() => setStatus("failure"));

    const getBody = useEffectEvent(() => props.body);

    const handleError = useEffectEvent((err: unknown) => {
        if (axios.isCancel(err)) return;

        if (err instanceof RateLimitedError) {
            setStatus("rate_limited");
            props.onError(new Error(translate("You have made too many requests")));
            triggerRateLimit(err.retryAfter);

            return;
        }

        console.error(err);
        props.onError(new Error(translate("There was an issue completing sign in process")));
        setStatus("failure");
    });

    const handleStatus = useEffectEvent((res: AppAuthenticatorSignInStatus | undefined) => {
        switch (res?.state) {
            case undefined:
                setStatus("success");
                props.onSuccess(res?.redirect);
                break;
            case "pending":
                setPoll((value) => value + 1);
                break;
            case "denied":
                setStatus("denied");
                break;
            case "rejected":
            case "expired":
                setStatus("not_approved");
                break;
            default:
                handleError(new Error(`unexpected state '${String(res?.state)}'`));
        }
    });

    useEffect(() => {
        if (status !== "starting") return;

        let active = true;

        initiateAppAuthenticatorSignIn(getBody(), getSignal())
            .then((res) => {
                if (!active) return;

                setChallenge(res);
                setStatus("pending");
            })
            .catch((err) => {
                if (active) handleError(err);
            });

        return () => {
            active = false;
        };
    }, [status, getSignal]);

    useEffect(() => {
        if (status !== "pending") return;

        let active = true;

        const timeout = setTimeout(() => {
            getAppAuthenticatorSignInStatus(getBody(), getSignal())
                .then((res) => {
                    if (active) handleStatus(res);
                })
                .catch((err) => {
                    if (active) handleError(err);
                });
        }, PollIntervalMs);

        return () => {
            active = false;
            clearTimeout(timeout);
        };
    }, [status, poll, getSignal]);

    const handleTryAgain = () => {
        setChallenge(undefined);
        setStatus("starting");
    };

    switch (status) {
        case "starting":
            return (
                <div className="flex w-full justify-center">
                    <Spinner size={32} />
                </div>
            );
        case "pending":
            return (
                <div className="flex w-full flex-col items-center gap-2">
                    <p
                        id="app-authenticator-number"
                        className="font-mono text-5xl font-bold tracking-[0.3em]"
                        aria-live="polite"
                    >
                        {challenge?.number}
                    </p>
                    <p>{translate("Open the Authelia app and enter this number")}</p>
                    <p id="app-authenticator-countdown" className="text-sm text-muted-foreground">
                        {translate("Expires in {{count}} seconds", { count: seconds })}
                    </p>
                </div>
            );
        case "success":
            return (
                <div className="flex w-full justify-center">
                    <SuccessIcon />
                </div>
            );
        default:
            return (
                <div className="flex w-full flex-col items-center gap-2">
                    <FailureIcon />
                    {status === "denied" ? (
                        <p id="app-authenticator-denied">{translate("Request denied on your device.")}</p>
                    ) : null}
                    {status === "not_approved" ? (
                        <p id="app-authenticator-not-approved">{translate("The request was not approved.")}</p>
                    ) : null}
                    <RateLimitCountdown secondsRemaining={secondsRemaining} />
                    {status === "not_approved" || status === "failure" ? (
                        <Button
                            id="app-authenticator-try-again"
                            variant={"ghost"}
                            color={"secondary"}
                            onClick={handleTryAgain}
                        >
                            {translate("Try again")}
                        </Button>
                    ) : null}
                </div>
            );
    }
};

export default AppAuthenticatorChallenge;
