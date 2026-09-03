import { Fragment, useCallback, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Separator } from "@components/UI/Separator";
import { Spinner } from "@components/UI/Spinner";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useAbortSignal } from "@hooks/Abort";
import { useQueryParam } from "@hooks/QueryParam";
import { OpenIDConnectProvider } from "@models/OpenIDConnectRelyingParty";
import { getOpenIDConnectProviders, postOpenIDConnectStart } from "@services/OpenIDConnectRelyingParty";

export interface Props {
    disabled: boolean;
    rememberMe: boolean;
}

const OpenIDConnectForm = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const requestMethod = useQueryParam(RequestMethod);
    const getSignal = useAbortSignal();

    const [providers, setProviders] = useState<OpenIDConnectProvider[]>([]);
    const [loading, setLoading] = useState<null | string>(null);

    useEffect(() => {
        const signal = getSignal();

        getOpenIDConnectProviders(signal)
            .then((values) => setProviders(values))
            .catch(() => setProviders([]));
    }, [getSignal]);

    const handleSignIn = useCallback(
        async (id: string) => {
            if (loading !== null) {
                return;
            }

            setLoading(id);

            try {
                const response = await postOpenIDConnectStart(
                    id,
                    {
                        keepMeLoggedIn: props.rememberMe,
                        requestMethod,
                        targetURL: redirectionURL,
                    },
                    getSignal(),
                );

                window.location.assign(response.authorization_url);
            } catch {
                setLoading(null);
            }
        },
        [loading, redirectionURL, requestMethod, props.rememberMe, getSignal],
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
                        id={`openid-connect-sign-in-button-${provider.id}`}
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

export default OpenIDConnectForm;
