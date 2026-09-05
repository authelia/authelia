import { Globe } from "lucide-react";
import { useTranslation } from "react-i18next";

import DecisionFormSection, { DecisionFormSectionItem } from "@views/ConsentPortal/OpenIDConnect/DecisionFormSection";

export interface Props {
    resource?: null | string[];
}

function DecisionFormResource({ resource }: Props) {
    const { t: translate } = useTranslation(["consent"]);

    if (!resource || resource.length === 0) {
        return null;
    }

    return (
        <DecisionFormSection
            id={"openid-consent-resource"}
            title={translate("Requested Resources")}
            description={translate("The protected resources the application intends to access")}
        >
            {resource.map((value: string) => (
                <DecisionFormSectionItem key={value} id={`resource-${value}`} icon={<Globe className="size-4" />}>
                    <span className="font-mono text-sm break-all">{value}</span>
                </DecisionFormSectionItem>
            ))}
        </DecisionFormSection>
    );
}

export default DecisionFormResource;
