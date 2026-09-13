// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { Alert, AlertDescription, AlertTitle } from "@components/UI/Alert";
import { IndexRoute } from "@constants/Routes";
import { useAbortSignal } from "@hooks/Abort";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { getExternalIdentityProviders } from "@services/ExternalIdentity";

export interface Props {
    provider: string;
}

const ExternalIdentityLinkNotice = function (props: Props) {
    const { t: translate } = useTranslation();

    const navigate = useRouterNavigate();
    const getSignal = useAbortSignal();

    const [name, setName] = useState(props.provider);

    useEffect(() => {
        getExternalIdentityProviders(getSignal())
            .then((providers) => {
                const provider = providers.find((value) => value.id === props.provider);

                if (provider) {
                    setName(provider.name);
                }
            })
            .catch(() => {});
    }, [props.provider, getSignal]);

    return (
        <Alert id="external-identity-link-notice">
            <AlertTitle>{translate("Link your {{name}} account", { name })}</AlertTitle>
            <AlertDescription>
                {translate(
                    "This {{name}} account is not linked yet, sign in to link it and you will then return to {{name}} to confirm the link",
                    { name },
                )}
                <button
                    id="external-identity-link-cancel"
                    type="button"
                    className="cursor-pointer text-primary underline-offset-4 hover:underline"
                    onClick={() => navigate(IndexRoute, false)}
                >
                    {translate("Use another sign in method")}
                </button>
            </AlertDescription>
        </Alert>
    );
};

export default ExternalIdentityLinkNotice;
