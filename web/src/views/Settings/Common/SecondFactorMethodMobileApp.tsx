// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback } from "react";

import { useNotifications } from "@contexts/NotificationsContext";
import { AppAuthenticatorSignInBody } from "@services/AppAuthenticator";
import AppAuthenticatorChallenge from "@views/LoginPortal/SecondFactor/AppAuthenticatorChallenge";

const body: AppAuthenticatorSignInBody = {};

export interface Props {
    onSecondFactorSuccess: () => void;
}

const SecondFactorMethodMobileApp = function (props: Props) {
    const { createErrorNotification } = useNotifications();
    const { onSecondFactorSuccess } = props;

    const handleSuccess = useCallback(() => {
        onSecondFactorSuccess();
    }, [onSecondFactorSuccess]);

    const handleError = useCallback(
        (err: Error) => {
            createErrorNotification(err.message);
        },
        [createErrorNotification],
    );

    return <AppAuthenticatorChallenge body={body} onSuccess={handleSuccess} onError={handleError} />;
};

export default SecondFactorMethodMobileApp;
