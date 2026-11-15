// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useEffect, useEffectEvent, useRef, useState, useSyncExternalStore } from "react";

import { WebAuthnAbortService, browserSupportsWebAuthnAutofill } from "@simplewebauthn/browser";
import axios from "axios";
import { useTranslation } from "react-i18next";

import PasskeyIcon from "@components/PasskeyIcon";
import RememberMeDialog from "@components/RememberMeDialog";
import { Button } from "@components/UI/Button";
import { Separator } from "@components/UI/Separator";
import { Spinner } from "@components/UI/Spinner";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useAbortSignal } from "@hooks/Abort";
import { useFlow } from "@hooks/Flow";
import { useQueryParam } from "@hooks/QueryParam";
import { useRememberMePrompt } from "@hooks/RememberMe";
import { AssertionResult, AssertionResultFailureString } from "@models/WebAuthn";
import { getWebAuthnPasskeyOptions, getWebAuthnResult, postWebAuthnPasskeyResponse } from "@services/WebAuthn";

const CONDITIONAL_REARM_MARGIN = 10000;
const CONDITIONAL_REARM_FALLBACK = 60000;
const CONDITIONAL_RETRY_BASE = 2000;
const CONDITIONAL_RETRY_MAX = 300000;

const subscribeVisibility = (callback: () => void) => {
    document.addEventListener("visibilitychange", callback);

    return () => document.removeEventListener("visibilitychange", callback);
};

const getVisibility = () => document.visibilityState === "visible";

const ignore = () => undefined;

export interface Props {
    disabled: boolean;
    rememberMe: boolean;

    onAuthenticationStart: () => void;
    onAuthenticationStop: () => void;
    onAuthenticationError: (_err: Error) => void;
    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
}

const PasskeyForm = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const requestMethod = useQueryParam(RequestMethod);
    const { flow, id: flowID, subflow } = useFlow();
    const getSignal = useAbortSignal();
    const visible = useSyncExternalStore(subscribeVisibility, getVisibility);

    const { dialogProps: rememberMeDialogProps, prompt: promptRememberMe } = useRememberMePrompt(props.rememberMe);

    const [loading, setLoading] = useState(false);
    const [conditionalGeneration, setConditionalGeneration] = useState(0);

    const loadingRef = useRef(false);
    const conditionalRef = useRef<AbortController | null>(null);
    const rearmTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
    const failuresRef = useRef(0);
    const pendingOptionsRef = useRef<null | Promise<unknown>>(null);

    const rearm = () => {
        setConditionalGeneration((generation) => generation + 1);
    };

    const clearRearm = () => {
        clearTimeout(rearmTimerRef.current);

        rearmTimerRef.current = undefined;
    };

    const scheduleRearm = (timeout: number = CONDITIONAL_REARM_FALLBACK) => {
        clearRearm();

        rearmTimerRef.current = setTimeout(
            () => {
                failuresRef.current = 0;

                rearm();
            },
            Math.max(timeout - CONDITIONAL_REARM_MARGIN, timeout / 2),
        );
    };

    const scheduleRetry = () => {
        clearRearm();

        const delay = Math.min(CONDITIONAL_RETRY_BASE * 2 ** failuresRef.current, CONDITIONAL_RETRY_MAX);

        failuresRef.current += 1;

        rearmTimerRef.current = setTimeout(rearm, delay);
    };

    const disarm = () => {
        clearRearm();

        conditionalRef.current?.abort();
        conditionalRef.current = null;
    };

    const fetchOptions = async (conditionalMediation: boolean) => {
        // The server keeps the challenge in the session, so an options request which has been superseded is still
        // allowed to settle before the next one is made. Aborting it would only abort it on the client, and it could
        // then still be saved after the next one.
        await pendingOptionsRef.current?.then(ignore, ignore);

        const request = getWebAuthnPasskeyOptions(getSignal(), conditionalMediation);

        pendingOptionsRef.current = request;

        try {
            return await request;
        } finally {
            if (pendingOptionsRef.current === request) pendingOptionsRef.current = null;
        }
    };

    const handleSignIn = async (conditional?: AbortSignal) => {
        if (loadingRef.current) return;

        const conditionalMediation = conditional !== undefined;

        let interactive = !conditionalMediation;
        let signal = conditional ?? getSignal();

        const startUI = () => {
            interactive = true;

            loadingRef.current = true;
            props.onAuthenticationStart();
            setLoading(true);
        };

        const stopUI = () => {
            if (!interactive) return;

            loadingRef.current = false;
            props.onAuthenticationStop();
            setLoading(false);
        };

        const fail = (message: string) => {
            if (!interactive) {
                scheduleRetry();

                return;
            }

            stopUI();

            props.onAuthenticationError(new Error(translate(message)));

            rearm();
        };

        if (!conditionalMediation) {
            disarm();
            startUI();
        }

        try {
            const optionsStatus = await fetchOptions(conditionalMediation);

            if (signal.aborted) return;

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                fail("Failed to initiate security key sign in process");

                return;
            }

            if (conditionalMediation) scheduleRearm(optionsStatus.options.timeout);

            const result = await getWebAuthnResult(optionsStatus.options, conditionalMediation);

            if (signal.aborted) return;

            if (result.result !== AssertionResult.Success) {
                fail(AssertionResultFailureString(result.result));

                return;
            }

            if (result.response == null) {
                fail("The browser did not respond with the expected attestation data");

                return;
            }

            if (conditionalMediation) {
                clearRearm();
                startUI();

                signal = getSignal();
            }

            const rememberMe = await promptRememberMe();

            if (signal.aborted) return;

            if (rememberMe === null) {
                stopUI();
                rearm();

                return;
            }

            const response = await postWebAuthnPasskeyResponse(
                result.response,
                rememberMe,
                redirectionURL,
                requestMethod,
                flowID,
                flow,
                subflow,
            );

            if (signal.aborted) {
                stopUI();

                return;
            }

            if (response.data.status === "OK" && response.status === 200) {
                stopUI();

                props.onAuthenticationSuccess(response.data.data ? response.data.data.redirect : undefined);

                return;
            }

            fail("The server rejected the security key");
        } catch (err) {
            if (axios.isCancel(err) || signal.aborted) {
                stopUI();

                return;
            }

            if (interactive) console.error(err);

            fail("Failed to initiate security key sign in process");
        }
    };

    const handleConditionalMediation = useEffectEvent(async (controller: AbortController) => {
        conditionalRef.current = controller;

        try {
            const supported = await browserSupportsWebAuthnAutofill();

            if (controller.signal.aborted || !supported) return;

            await handleSignIn(controller.signal);
        } catch (err) {
            if (axios.isCancel(err)) return;

            console.error(err);
        }
    });

    const handleConditionalMediationStop = useEffectEvent((controller: AbortController) => {
        clearRearm();

        controller.abort();

        if (!loadingRef.current) WebAuthnAbortService.cancelCeremony();
    });

    useEffect(() => {
        if (!visible) return;

        const controller = new AbortController();

        void handleConditionalMediation(controller);

        return () => handleConditionalMediationStop(controller);
    }, [conditionalGeneration, visible]);

    return (
        <Fragment>
            <div className="w-full">
                <div className="relative flex items-center py-2">
                    <Separator className="flex-1" />
                    <span className="px-3 text-sm uppercase text-muted-foreground">{translate("or")}</span>
                    <Separator className="flex-1" />
                </div>
            </div>
            <div className="w-full">
                <Button
                    id="passkey-sign-in-button"
                    variant="default"
                    className="w-full"
                    onClick={() => void handleSignIn()}
                    disabled={props.disabled}
                >
                    <PasskeyIcon />
                    {translate("Sign in with a passkey")}
                    {loading ? <Spinner size={20} className="ml-2 h-5 w-5" /> : null}
                </Button>
            </div>
            <RememberMeDialog {...rememberMeDialogProps} />
        </Fragment>
    );
};

export default PasskeyForm;
