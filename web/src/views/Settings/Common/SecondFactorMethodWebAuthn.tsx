// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useEffect } from "react";

import WebAuthnTryIcon from "@components/WebAuthnTryIcon";
import { useWebAuthnAssertion } from "@hooks/WebAuthnAssertion";
import { getWebAuthnOptions, postWebAuthnResponse } from "@services/WebAuthn";

export interface Props {
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodWebAuthn = function (props: Props) {
    const { ceremony, start, state } = useWebAuthnAssertion({
        getOptions: getWebAuthnOptions,
        onError: (err) => console.error(err),
        onSuccess: () => props.onSecondFactorSuccess(),
        postResponse: (response, signal) =>
            postWebAuthnResponse(response, undefined, undefined, undefined, undefined, undefined, signal),
    });

    useEffect(() => {
        start().catch(console.error);
    }, [start]);

    const handleRetry = () => {
        start().catch(console.error);
    };

    return <WebAuthnTryIcon ceremony={ceremony} onRetryClick={handleRetry} webauthnTouchState={state} />;
};

export default SecondFactorMethodWebAuthn;
