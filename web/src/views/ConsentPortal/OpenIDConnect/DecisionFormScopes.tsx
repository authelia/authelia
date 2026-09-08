import { useTranslation } from "react-i18next";

import { ScopeAvatar } from "@components/OpenIDConnect";
import { formatScope } from "@services/ConsentOpenIDConnect";
import DecisionFormSection, { DecisionFormSectionItem } from "@views/ConsentPortal/OpenIDConnect/DecisionFormSection";

export interface Props {
    scopes?: null | string[];
}

function DecisionFormScopes({ scopes }: Props) {
    const { t: translate } = useTranslation(["consent"]);

    if (!scopes || scopes.length === 0) {
        return null;
    }

    return (
        <DecisionFormSection
            id={"openid-consent-scopes"}
            title={translate("Requested Permissions")}
            description={translate("The actions the application will be allowed to perform on your behalf")}
        >
            {scopes.map((scope: string) => {
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
            })}
        </DecisionFormSection>
    );
}

export default DecisionFormScopes;
