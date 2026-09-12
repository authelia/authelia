// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useEffect, useEffectEvent, useRef, useState } from "react";

import { browserSupportsWebAuthnAutofill } from "@simplewebauthn/browser";
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

    const { dialogProps: rememberMeDialogProps, prompt: promptRememberMe } = useRememberMePrompt(props.rememberMe);

    const [loading, setLoading] = useState(false);

    const unmountedRef = useRef(false);

    const handleSignIn = async (conditionalMediation: boolean) => {
        if (loading) return;

        // A conditional ceremony waits in the background for the user to pick a credential from the browser's
        // autofill, so until they do the form must neither present itself as busy nor report failures: the user never
        // asked for it, and the explicit button remains available to them.
        let interactive = !conditionalMediation;

        const startUI = () => {
            interactive = true;

            props.onAuthenticationStart();
            setLoading(true);
        };

        const stopUI = () => {
            // A ceremony only ever tears down the busy state it put up itself: a conditional ceremony is superseded by
            // an explicit one, and must not clear the busy state belonging to the ceremony that replaced it.
            if (!interactive) return;

            props.onAuthenticationStop();
            setLoading(false);
        };

        const fail = (message: string) => {
            stopUI();

            if (!interactive) return;

            props.onAuthenticationError(new Error(translate(message)));
        };

        if (!conditionalMediation) startUI();

        const signal = getSignal();

        try {
            const optionsStatus = await getWebAuthnPasskeyOptions(signal, conditionalMediation);

            if (signal.aborted) return;

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                fail("Failed to initiate security key sign in process");

                return;
            }

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

            // The user has picked a credential, so from here the ceremony is theirs regardless of how it started. The
            // remember me choice is only put to them at this point, so that a conditionally mediated sign in asks at
            // the same point an explicit one does rather than ahead of any user intent.
            if (conditionalMediation) startUI();

            const rememberMe = await promptRememberMe();

            if (signal.aborted) return;

            const response = await postWebAuthnPasskeyResponse(
                result.response,
                rememberMe,
                redirectionURL,
                requestMethod,
                flowID,
                flow,
                subflow,
                signal,
            );

            stopUI();

            if (response.data.status === "OK" && response.status === 200) {
                props.onAuthenticationSuccess(response.data.data ? response.data.data.redirect : undefined);

                return;
            }

            props.onAuthenticationError(new Error(translate("The server rejected the security key")));
        } catch (err) {
            stopUI();

            if (axios.isCancel(err) || !interactive) return;

            console.error(err);

            props.onAuthenticationError(new Error(translate("Failed to initiate security key sign in process")));
        }
    };

    const handleConditionalMediation = useEffectEvent(async () => {
        try {
            const supported = await browserSupportsWebAuthnAutofill();

            if (unmountedRef.current || !supported) return;

            await handleSignIn(true);
        } catch (err) {
            if (axios.isCancel(err)) return;

            console.error(err);
        }
    });

    useEffect(() => {
        unmountedRef.current = false;

        void handleConditionalMediation();

        return () => {
            unmountedRef.current = true;
        };
    }, []);

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
                    onClick={() => void handleSignIn(false)}
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
