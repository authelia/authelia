import React, { useCallback, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { useIsMountedRef } from "@hooks/Mounted";
import { AssertionResult, AssertionResultFailureString, WebAuthnTouchState } from "@models/WebAuthn";
import { getWebAuthnOptions, getWebAuthnResult, postWebAuthnResponse } from "@services/WebAuthn";

export interface Props {
    closing: boolean;
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodWebAuthn = function (props: Props) {
    const [state, setState] = useState(WebAuthnTouchState.WaitTouch);
    const [started, setStarted] = useState(false);
    const mounted = useIsMountedRef();
    const { t: translate } = useTranslation();

    const handleRetry = () => {
        setState(WebAuthnTouchState.WaitTouch);
        setStarted(false);
    };

    const handleStart = useCallback(async () => {
        setStarted(true);

        try {
            const optionsStatus = await getWebAuthnOptions();

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                if (!mounted.current) return;
                setState(WebAuthnTouchState.Failure);
                console.error(new Error(translate("Failed to initiate security key sign in process")));

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (result.result !== AssertionResult.Success) {
                if (mounted.current) {
                    setState(WebAuthnTouchState.Failure);
                    console.error(new Error(translate(AssertionResultFailureString(result.result))));
                }
                return;
            }

            if (result.response == null) {
                console.error(new Error(translate("The browser did not respond with the expected attestation data")));
                if (mounted.current) {
                    setState(WebAuthnTouchState.Failure);
                }

                return;
            }

            if (!mounted.current) return;

            setState(WebAuthnTouchState.InProgress);

            const response = await postWebAuthnResponse(result.response);

            if (response.data.status === "OK" && response.status === 200) {
                if (!mounted.current) return;
                props.onSecondFactorSuccess();
                return;
            }

            if (!mounted.current) return;

            console.error(new Error(translate("The server rejected the security key")));
            setState(WebAuthnTouchState.Failure);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            setState(WebAuthnTouchState.Failure);
        }
    }, [mounted, props, translate]);

    useEffect(() => {
        if (started) return;

        handleStart().catch(console.error);
    }, [handleStart, started]);

    return <WebAuthnTryIcon onRetryClick={handleRetry} webauthnTouchState={state} />;
};

export default SecondFactorMethodWebAuthn;
