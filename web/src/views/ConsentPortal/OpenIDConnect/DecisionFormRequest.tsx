// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useState } from "react";

import { ChevronDown } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Card, CardContent, CardHeader } from "@components/UI/Card";
import { Separator } from "@components/UI/Separator";
import { ConsentGetResponseBody } from "@services/ConsentOpenIDConnect";
import DecisionFormAudience from "@views/ConsentPortal/OpenIDConnect/DecisionFormAudience";
import DecisionFormClaims from "@views/ConsentPortal/OpenIDConnect/DecisionFormClaims";
import DecisionFormClient from "@views/ConsentPortal/OpenIDConnect/DecisionFormClient";
import DecisionFormResource from "@views/ConsentPortal/OpenIDConnect/DecisionFormResource";
import DecisionFormScopes from "@views/ConsentPortal/OpenIDConnect/DecisionFormScopes";

export interface Props {
    response: ConsentGetResponseBody;
    claims: string[];
    onChangeClaims: (_claims: string[]) => void;
    collapsible?: boolean;
}

function DecisionFormRequest({ claims, collapsible, onChangeClaims, response }: Props) {
    const { t: translate } = useTranslation(["consent"]);

    const [open, setOpen] = useState(false);

    const hasScopes = (response.scopes?.length ?? 0) > 0;

    const sections = (
        <CardContent className="flex flex-col gap-5 px-4 py-5">
            <DecisionFormScopes scopes={response.scopes} headless />
            <DecisionFormClaims
                claims={response.claims}
                checked={claims}
                essential_claims={response.essential_claims}
                onChangeChecked={onChangeClaims}
            />
            <DecisionFormAudience audience={response.audience} />
            <DecisionFormResource resource={response.resource} />
        </CardContent>
    );

    const empty =
        (response.scopes?.length ?? 0) === 0 &&
        (response.claims?.length ?? 0) === 0 &&
        (response.essential_claims?.length ?? 0) === 0 &&
        (response.audience?.length ?? 0) === 0 &&
        (response.resource?.length ?? 0) === 0;

    return (
        <div className="flex w-full flex-col gap-4">
            <DecisionFormClient
                client_id={response.client_id}
                client_description={response.client_description}
                client_logo_uri={response.client_logo_uri}
            />
            <Card className="gap-0 overflow-hidden py-0">
                {hasScopes ? (
                    <CardHeader className="px-4 pt-4 pb-2 text-left">
                        <h3 className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">
                            {translate("Requested Permissions")}
                        </h3>
                        <p className="mt-0.5 text-xs text-muted-foreground">
                            {translate("The actions the application will be allowed to perform on your behalf")}
                        </p>
                    </CardHeader>
                ) : null}
                {collapsible ? (
                    empty ? null : (
                        <div>
                            {hasScopes ? <Separator /> : null}
                            <button
                                type="button"
                                id={"openid-consent-request-disclosure"}
                                className="flex w-full items-center gap-2 px-4 py-2.5 text-left text-xs font-medium text-muted-foreground hover:text-foreground"
                                aria-expanded={open}
                                aria-controls={"openid-consent-request-details"}
                                onClick={() => setOpen((previous) => !previous)}
                            >
                                <ChevronDown className={`size-4 transition-transform ${open ? "rotate-180" : ""}`} />
                                {open ? translate("Hide request details") : translate("Show request details")}
                            </button>
                            {open ? <div id={"openid-consent-request-details"}>{sections}</div> : null}
                        </div>
                    )
                ) : (
                    <div>
                        {hasScopes ? <Separator /> : null}
                        {sections}
                    </div>
                )}
            </Card>
        </div>
    );
}

export default DecisionFormRequest;
