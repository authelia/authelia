// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Fragment, useCallback, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Separator } from "@components/UI/Separator";
import { Spinner } from "@components/UI/Spinner";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useNotifications } from "@contexts/NotificationsContext";
import { useAbortSignal } from "@hooks/Abort";
import { useFlow } from "@hooks/Flow";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useQueryParam } from "@hooks/QueryParam";
import { ExternalIdentityProvider } from "@models/ExternalIdentity";
import { getExternalIdentityProviders, postExternalIdentityStart } from "@services/ExternalIdentity";

export interface Props {
    disabled: boolean;
    rememberMe: boolean;

    onLoadingChange?: (loading: boolean) => void;
}

const ExternalIdentityForm = function (props: Props) {
    const { t: translate } = useTranslation();
    const { createErrorNotification } = useNotifications();

    const redirectionURL = useQueryParam(RedirectionURL);
    const requestMethod = useQueryParam(RequestMethod);
    const { flow, id: flowID, subflow } = useFlow();
    const userCode = useUserCode();
    const getSignal = useAbortSignal();

    const [providers, setProviders] = useState<ExternalIdentityProvider[]>([]);
    const [loading, setLoading] = useState<null | string>(null);

    useEffect(() => {
        const signal = getSignal();

        getExternalIdentityProviders(signal)
            .then((values) => setProviders(values))
            .catch(() => setProviders([]));
    }, [getSignal]);

    const handleSignIn = useCallback(
        async (id: string) => {
            if (loading !== null) {
                return;
            }

            setLoading(id);
            props.onLoadingChange?.(true);

            const signal = getSignal();

            try {
                const response = await postExternalIdentityStart(
                    id,
                    {
                        flow,
                        flowID,
                        keepMeLoggedIn: props.rememberMe,
                        requestMethod,
                        subflow,
                        targetURL: redirectionURL,
                        userCode,
                    },
                    signal,
                );

                window.location.assign(response.authorization_url);
            } catch {
                if (signal.aborted) return;

                setLoading(null);
                props.onLoadingChange?.(false);
                createErrorNotification(translate("There was an issue signing in with the external provider"));
            }
        },
        [
            loading,
            redirectionURL,
            requestMethod,
            flow,
            flowID,
            subflow,
            userCode,
            props,
            getSignal,
            createErrorNotification,
            translate,
        ],
    );

    if (providers.length === 0) {
        return null;
    }

    return (
        <Fragment>
            <div className="w-full">
                <div className="relative flex items-center py-2">
                    <Separator className="flex-1" />
                    <span className="px-3 text-sm uppercase text-muted-foreground">{translate("or")}</span>
                    <Separator className="flex-1" />
                </div>
            </div>
            {providers.map((provider) => (
                <div className="w-full" key={provider.id}>
                    <Button
                        id={`external-identity-sign-in-button-${provider.id}`}
                        variant="outline"
                        className="w-full"
                        disabled={props.disabled || loading !== null}
                        onClick={() => handleSignIn(provider.id)}
                    >
                        {translate("Sign in with {{name}}", { name: provider.name })}
                        {loading === provider.id ? <Spinner size={20} className="ml-2 h-5 w-5" /> : null}
                    </Button>
                </div>
            ))}
        </Fragment>
    );
};

export default ExternalIdentityForm;
