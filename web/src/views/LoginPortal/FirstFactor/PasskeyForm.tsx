import { Fragment, useCallback, useRef, useState } from "react";

import { useTranslation } from "react-i18next";

import PasskeyIcon from "@components/PasskeyIcon";
import { Button } from "@components/UI/Button";
import { Separator } from "@components/UI/Separator";
import { Spinner } from "@components/UI/Spinner";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useIsMountedRef } from "@hooks/Mounted";
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
    const mounted = useIsMountedRef();

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

    const handleSignIn = useCallback(async () => {
        if (!mounted || loading) {
            return;
        }

        handleAuthenticationStart();

        try {
            const optionsStatus = await getWebAuthnPasskeyOptions();

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                handleAuthenticationStop();
                onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (result.result !== AssertionResult.Success) {
                if (!mounted.current) return;

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

            if (!mounted.current) return;

            const response = await postWebAuthnPasskeyResponse(
                result.response,
                props.rememberMe,
                redirectionURL,
                requestMethod,
                flowID,
                flow,
                subflow,
            );

            handleAuthenticationStop();

            if (response.data.status === "OK" && response.status === 200) {
                props.onAuthenticationSuccess(response.data.data ? response.data.data.redirect : undefined);
                return;
            }

            if (!mounted.current) return;

            onSignInErrorCallback(new Error(translate("The server rejected the security key")));
        } catch (err) {
            handleAuthenticationStop();

            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));
        }
    }, [
        mounted,
        loading,
        handleAuthenticationStart,
        props,
        redirectionURL,
        requestMethod,
        flowID,
        flow,
        subflow,
        handleAuthenticationStop,
        onSignInErrorCallback,
        translate,
    ]);

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
                    {loading ? <Spinner className="ml-2 h-5 w-5" /> : null}
                </Button>
            </div>
        </Fragment>
    );
};

export default PasskeyForm;
