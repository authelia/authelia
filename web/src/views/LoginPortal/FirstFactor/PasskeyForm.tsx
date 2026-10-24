// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useCallback, useEffect, useRef, useState } from "react";

import { AuthenticationResponseJSON } from "@simplewebauthn/browser";
import axios from "axios";
import { useTranslation } from "react-i18next";

import PasskeyIcon from "@components/PasskeyIcon";
import { Button } from "@components/UI/Button";
import { Separator } from "@components/UI/Separator";
import { Spinner } from "@components/UI/Spinner";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useAbortSignal } from "@hooks/Abort";
import { useFlow } from "@hooks/Flow";
import { useQueryParam } from "@hooks/QueryParam";
import { AssertionResult, AssertionResultFailureString } from "@models/WebAuthn";
import {
    cancelWebAuthnCeremony,
    getWebAuthnAutofillSupported,
    getWebAuthnPasskeyOptions,
    getWebAuthnResult,
    postWebAuthnPasskeyResponse,
} from "@services/WebAuthn";

// The conditional mediation challenge is refreshed this long before the server side challenge expires, as browsers
// ignore the timeout for conditional mediation requests but the server enforces it.
const ConditionalMediationRefreshMargin = 5000;
const ConditionalMediationRefreshMinimum = 15000;
const ConditionalMediationDefaultTimeout = 60000;
const ConditionalMediationRetryDelay = 1000;

export interface Props {
    disabled: boolean;
    rememberMe: boolean;
    conditionalMediation: boolean;

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

    const [loading, setLoading] = useState(false);

    const onSignInErrorCallback = useRef(props.onAuthenticationError).current;

    const handleAuthenticationStart = useCallback(() => {
        props.onAuthenticationStart();
        setLoading(true);
    }, [props]);

    const handleAuthenticationStop = useCallback(() => {
        props.onAuthenticationStop();
        setLoading(false);
    }, [props]);

    // Tracks if a conditional mediation ceremony is currently pending so it is only ever cancelled by this component when
    // it is the active ceremony, and never cancels a modal ceremony started by the sign in button.
    const conditionalPendingRef = useRef(false);

    // Tracks the conditional mediation refresh or retry timer so it is cleared whenever conditional mediation is
    // cancelled, including when no ceremony is pending, as it would otherwise replace the challenge of a modal ceremony.
    const conditionalTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

    const cancelConditionalMediation = useCallback(() => {
        clearTimeout(conditionalTimerRef.current);
        conditionalTimerRef.current = undefined;

        if (!conditionalPendingRef.current) return;

        conditionalPendingRef.current = false;
        cancelWebAuthnCeremony();
    }, []);

    const handleAssertionResponse = useCallback(
        async (assertion: AuthenticationResponseJSON, signal: AbortSignal) => {
            const response = await postWebAuthnPasskeyResponse(
                assertion,
                props.rememberMe,
                redirectionURL,
                requestMethod,
                flowID,
                flow,
                subflow,
                signal,
            );

            handleAuthenticationStop();

            if (response.data.status === "OK" && response.status === 200) {
                props.onAuthenticationSuccess(response.data.data ? response.data.data.redirect : undefined);
                return;
            }

            onSignInErrorCallback(new Error(translate("The server rejected the security key")));
        },
        [
            props,
            redirectionURL,
            requestMethod,
            flowID,
            flow,
            subflow,
            handleAuthenticationStop,
            onSignInErrorCallback,
            translate,
        ],
    );

    const handleSignIn = useCallback(async () => {
        if (loading) {
            return;
        }

        cancelConditionalMediation();

        handleAuthenticationStart();

        const signal = getSignal();

        try {
            const optionsStatus = await getWebAuthnPasskeyOptions(signal);

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                handleAuthenticationStop();
                onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (signal.aborted) return;

            if (result.result !== AssertionResult.Success) {
                handleAuthenticationStop();

                onSignInErrorCallback(new Error(translate(AssertionResultFailureString(result.result))));

                return;
            }

            if (result.response == null) {
                onSignInErrorCallback(
                    new Error(translate("The browser did not respond with the expected attestation data")),
                );
                handleAuthenticationStop();

                return;
            }

            await handleAssertionResponse(result.response, signal);
        } catch (err) {
            handleAuthenticationStop();

            if (axios.isCancel(err)) return;
            console.error(err);
            onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));
        }
    }, [
        getSignal,
        loading,
        cancelConditionalMediation,
        handleAuthenticationStart,
        handleAssertionResponse,
        handleAuthenticationStop,
        onSignInErrorCallback,
        translate,
    ]);

    const handleConditionalAssertion = useCallback(
        async (assertion: AuthenticationResponseJSON, signal: AbortSignal) => {
            handleAuthenticationStart();

            try {
                await handleAssertionResponse(assertion, signal);
            } catch (err) {
                handleAuthenticationStop();

                if (axios.isCancel(err)) return;
                console.error(err);
                onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));
            }
        },
        [
            handleAuthenticationStart,
            handleAssertionResponse,
            handleAuthenticationStop,
            onSignInErrorCallback,
            translate,
        ],
    );

    // The latest handler is kept in a ref so the conditional mediation effect does not restart on every render.
    const conditionalAssertionHandlerRef = useRef(handleConditionalAssertion);

    useEffect(() => {
        conditionalAssertionHandlerRef.current = handleConditionalAssertion;
    }, [handleConditionalAssertion]);

    useEffect(() => {
        if (!props.conditionalMediation || props.disabled) return;

        let cancelled = false;
        let generation = 0;

        const start = async () => {
            const current = ++generation;
            const stale = () => cancelled || current !== generation;

            try {
                if (!(await getWebAuthnAutofillSupported()) || stale()) return;

                const signal = getSignal();

                const optionsStatus = await getWebAuthnPasskeyOptions(signal, "conditional");

                if (
                    stale() ||
                    optionsStatus.status !== 200 ||
                    optionsStatus.options == null ||
                    optionsStatus.mediation !== "conditional"
                ) {
                    return;
                }

                const timeout = optionsStatus.options.timeout ?? ConditionalMediationDefaultTimeout;
                const refresh = setTimeout(
                    () => {
                        cancelConditionalMediation();
                        void start();
                    },
                    Math.max(
                        timeout - ConditionalMediationRefreshMargin,
                        Math.min(timeout / 2, ConditionalMediationRefreshMinimum),
                    ),
                );

                conditionalTimerRef.current = refresh;
                conditionalPendingRef.current = true;

                const result = await getWebAuthnResult(optionsStatus.options, true);

                // Never refresh once the ceremony has finished, so the challenge in the session can't be replaced
                // between the assertion and its validation.
                clearTimeout(refresh);

                if (current === generation) {
                    conditionalPendingRef.current = false;
                }

                if (stale() || result.result === AssertionResult.FailureAborted) return;

                if (result.result !== AssertionResult.Success || result.response == null) {
                    if (result.result === AssertionResult.FailureUserConsent) {
                        // The user dismissed the request, quietly start a new one so the Passkey is still offered.
                        conditionalTimerRef.current = setTimeout(() => void start(), ConditionalMediationRetryDelay);
                    }

                    return;
                }

                await conditionalAssertionHandlerRef.current(result.response, signal);
            } catch (err) {
                // Failures before an assertion exists are not shown, the sign in button remains available.
                if (!axios.isCancel(err)) console.error(err);
            }
        };

        void start();

        return () => {
            cancelled = true;
            clearTimeout(conditionalTimerRef.current);
            cancelConditionalMediation();
        };
    }, [props.conditionalMediation, props.disabled, getSignal, cancelConditionalMediation]);

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
                    onClick={handleSignIn}
                    disabled={props.disabled}
                >
                    <PasskeyIcon />
                    {translate("Sign in with a passkey")}
                    {loading ? <Spinner size={20} className="ml-2 h-5 w-5" /> : null}
                </Button>
            </div>
        </Fragment>
    );
};

export default PasskeyForm;
