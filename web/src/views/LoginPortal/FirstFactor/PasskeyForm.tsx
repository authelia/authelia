// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useState } from "react";

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

    const [loading, setLoading] = useState(false);

    const handleSignIn = async () => {
        if (loading) return;

        const startUI = () => {
            props.onAuthenticationStart();
            setLoading(true);
        };

        const stopUI = () => {
            props.onAuthenticationStop();
            setLoading(false);
        };

        const fail = (message: string) => {
            stopUI();
            props.onAuthenticationError(new Error(translate(message)));
        };

        startUI();

        const signal = getSignal();

        try {
            const optionsStatus = await getWebAuthnPasskeyOptions(signal);

            if (signal.aborted) return;

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                fail("Failed to initiate security key sign in process");

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (signal.aborted) return;

            if (result.result !== AssertionResult.Success) {
                fail(AssertionResultFailureString(result.result));

                return;
            }

            if (result.response == null) {
                fail("The browser did not respond with the expected attestation data");

                return;
            }

            const response = await postWebAuthnPasskeyResponse(
                result.response,
                props.rememberMe,
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

            if (axios.isCancel(err)) return;

            console.error(err);

            props.onAuthenticationError(new Error(translate("Failed to initiate security key sign in process")));
        }
    };

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
        </Fragment>
    );
};

export default PasskeyForm;
