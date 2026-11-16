// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useCallback, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import PasskeyIcon from "@components/PasskeyIcon";
import { Button } from "@components/UI/Button";
import { Separator } from "@components/UI/Separator";
import { Spinner } from "@components/UI/Spinner";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useQueryParam } from "@hooks/QueryParam";
import { useWebAuthnAssertion } from "@hooks/WebAuthnAssertion";
import { getWebAuthnPasskeyOptions, postWebAuthnPasskeyResponse } from "@services/WebAuthn";

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

    const [loading, setLoading] = useState(false);
    const activeRef = useRef(false);

    const handleAuthenticationStop = useCallback(() => {
        if (!activeRef.current) return;

        activeRef.current = false;
        props.onAuthenticationStop();
        setLoading(false);
    }, [props]);

    const { start } = useWebAuthnAssertion({
        getOptions: getWebAuthnPasskeyOptions,
        onError: (err) => {
            handleAuthenticationStop();
            props.onAuthenticationError(err);
        },
        onSuccess: (redirectURL) => {
            handleAuthenticationStop();
            props.onAuthenticationSuccess(redirectURL);
        },
        postResponse: (response, signal) =>
            postWebAuthnPasskeyResponse(
                response,
                props.rememberMe,
                redirectionURL,
                requestMethod,
                flowID,
                flow,
                subflow,
                signal,
            ),
    });

    const handleSignIn = useCallback(async () => {
        if (activeRef.current) {
            return;
        }

        activeRef.current = true;
        props.onAuthenticationStart();
        setLoading(true);

        await start();

        handleAuthenticationStop();
    }, [handleAuthenticationStop, props, start]);

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
                    aria-busy={loading}
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
