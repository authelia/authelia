import { Server } from "lucide-react";
import { useTranslation } from "react-i18next";

import DecisionFormSection, { DecisionFormSectionItem } from "@views/ConsentPortal/OpenIDConnect/DecisionFormSection";

export interface Props {
    audience?: null | string[];
}

function DecisionFormAudience({ audience }: Props) {
    const { t: translate } = useTranslation(["consent"]);

    if (!audience || audience.length === 0) {
        return null;
    }

    return (
        <DecisionFormSection
            id={"openid-consent-audience"}
            title={translate("Token Audience")}
            description={translate("The services the issued tokens will be valid for")}
        >
            {audience.map((value: string) => (
                <DecisionFormSectionItem key={value} id={`audience-${value}`} icon={<Server className="size-4" />}>
                    <span className="font-mono text-sm break-all">{value}</span>
                </DecisionFormSectionItem>
            ))}
        </DecisionFormSection>
    );
}

export default DecisionFormAudience;
