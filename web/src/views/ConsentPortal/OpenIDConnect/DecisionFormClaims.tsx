import { useCallback, useMemo, useState } from "react";

import { useTranslation } from "react-i18next";

import { Checkbox } from "@components/UI/Checkbox";
import { Label } from "@components/UI/Label";
import { formatClaim } from "@services/ConsentOpenIDConnect";
import DecisionFormSection, { DecisionFormSectionItem } from "@views/ConsentPortal/OpenIDConnect/DecisionFormSection";

export interface Props {
    onChangeChecked: (_claims: string[]) => void;
    claims: null | string[];
    essential_claims: null | string[];
}

function DecisionFormClaims({ claims, essential_claims, onChangeChecked }: Props) {
    const { t: translate } = useTranslation(["consent"]);

    const [availableClaims] = useState(() => claims || []);
    const checked = useMemo(() => claims || [], [claims]);

    const handleClaimCheckboxOnChange = (claim: string) => {
        const checking = !checked.includes(claim);

        if (checking) {
            onChangeChecked([...checked, claim]);
        } else {
            onChangeChecked(checked.filter((value) => value !== claim));
        }
    };

    const claimChecked = useCallback(
        (claim: string) => {
            return checked.includes(claim);
        },
        [checked],
    );

    const label = useCallback(
        (claim: string) => formatClaim(translate(`claims.${claim}`, { nsSeparator: false }), claim),
        [translate],
    );

    const identifier = useCallback(
        (claim: string) => (label(claim).toLowerCase() === claim.toLowerCase() ? null : claim),
        [label],
    );

    const hasClaims = (essential_claims && essential_claims.length > 0) || availableClaims.length > 0;

    if (!hasClaims) {
        return null;
    }

    return (
        <DecisionFormSection
            id={"openid-consent-claims"}
            title={translate("Information Shared")}
            description={translate("The information about you the application will receive")}
        >
            {essential_claims?.map((claim: string) => (
                <DecisionFormSectionItem
                    key={`${claim}-essential`}
                    icon={<Checkbox id={`claim-${claim}-essential`} disabled checked />}
                    identifier={identifier(claim)}
                    actions={
                        <span className="rounded-full bg-muted px-2 py-0.5 text-[0.625rem] font-medium tracking-wide text-muted-foreground uppercase">
                            {translate("Required")}
                        </span>
                    }
                >
                    <Label htmlFor={`claim-${claim}-essential`} className="text-sm">
                        {label(claim)}
                    </Label>
                </DecisionFormSectionItem>
            ))}
            {availableClaims.map((claim: string) => (
                <DecisionFormSectionItem
                    key={claim}
                    icon={
                        <Checkbox
                            id={`claim-${claim}`}
                            checked={claimChecked(claim)}
                            onCheckedChange={() => handleClaimCheckboxOnChange(claim)}
                        />
                    }
                    identifier={identifier(claim)}
                >
                    <Label htmlFor={`claim-${claim}`} className="text-sm">
                        {label(claim)}
                    </Label>
                </DecisionFormSectionItem>
            ))}
        </DecisionFormSection>
    );
}

export default DecisionFormClaims;
