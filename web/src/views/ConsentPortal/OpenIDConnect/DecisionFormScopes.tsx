// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useTranslation } from "react-i18next";

import { ScopeAvatar } from "@components/OpenIDConnect";
import { ItemGroup } from "@components/UI/Item";
import { formatScope } from "@services/ConsentOpenIDConnect";
import DecisionFormSection, { DecisionFormSectionItem } from "@views/ConsentPortal/OpenIDConnect/DecisionFormSection";

export interface Props {
    scopes?: null | string[];
    headless?: boolean;
}

function DecisionFormScopes({ headless, scopes }: Props) {
    const { t: translate } = useTranslation(["consent"]);

    if (!scopes || scopes.length === 0) {
        return null;
    }

    const items = scopes.map((scope: string) => {
        const label = formatScope(translate(`scopes.${scope}`, { nsSeparator: false }), scope);

        return (
            <DecisionFormSectionItem
                key={scope}
                id={`scope-${scope}`}
                icon={ScopeAvatar(scope)}
                identifier={label.toLowerCase() === scope.toLowerCase() ? null : scope}
            >
                <span className="text-sm">{label}</span>
            </DecisionFormSectionItem>
        );
    });

    if (headless) {
        return (
            <ItemGroup id={"openid-consent-scopes"} className="gap-1">
                {items}
            </ItemGroup>
        );
    }

    return (
        <DecisionFormSection
            id={"openid-consent-scopes"}
            title={translate("Requested Permissions")}
            description={translate("The actions the application will be allowed to perform on your behalf")}
        >
            {items}
        </DecisionFormSection>
    );
}

export default DecisionFormScopes;
