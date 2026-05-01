// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useState } from "react";

import { AppWindow } from "lucide-react";
import { useTranslation } from "react-i18next";

export interface Props {
    client_id: string;
    client_description: string;
    client_logo_uri?: string;
}

function DecisionFormClient({ client_description, client_id, client_logo_uri }: Props) {
    const { t: translate } = useTranslation(["consent"]);

    const [logoLoadError, setLogoLoadError] = useState(false);

    const named = client_description !== "";

    return (
        <div className="flex w-full items-center gap-3 text-left">
            <span className="flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-lg bg-muted text-muted-foreground">
                {client_logo_uri && !logoLoadError ? (
                    <img
                        id={"openid-consent-client-logo"}
                        src={client_logo_uri}
                        alt={named ? client_description : client_id}
                        onError={() => setLogoLoadError(true)}
                        className="size-full object-contain"
                    />
                ) : (
                    <AppWindow className="size-5" />
                )}
            </span>
            <span className="min-w-0 flex-1">
                <span
                    id={"openid-consent-client-name"}
                    data-testid={"openid-consent-client-name"}
                    className="block truncate font-semibold"
                >
                    {named ? client_description : client_id}
                </span>
                {named ? (
                    <span
                        id={"openid-consent-client-id"}
                        className="block truncate font-mono text-xs text-muted-foreground"
                        title={translate("Client ID", { client_id }) || `Client ID: ${client_id}`}
                    >
                        {client_id}
                    </span>
                ) : null}
            </span>
        </div>
    );
}

export default DecisionFormClient;
