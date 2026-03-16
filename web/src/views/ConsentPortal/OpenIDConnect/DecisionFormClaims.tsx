import { FC, Fragment, useCallback, useMemo, useState } from "react";

import { useTranslation } from "react-i18next";

import { Checkbox } from "@components/UI/Checkbox";
import { Label } from "@components/UI/Label";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { formatClaim } from "@services/ConsentOpenIDConnect";

export interface Props {
    onChangeChecked: (_claims: string[]) => void;
    claims: null | string[];
    essential_claims: null | string[];
}

const DecisionFormClaims: FC<Props> = ({ claims, essential_claims, onChangeChecked }: Props) => {
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

    const hasClaims = (essential_claims && essential_claims.length > 0) || availableClaims.length > 0;

    return (
        <Fragment>
            {hasClaims ? (
                <ul className="my-3 list-none rounded-md bg-card p-2">
                    {essential_claims?.map((claim: string) => (
                        <TooltipProvider key={`${claim}-essential`}>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <li className="flex items-center gap-2 px-2 py-1">
                                        <Checkbox id={`claim-${claim}-essential`} disabled checked />
                                        <Label htmlFor={`claim-${claim}-essential`}>
                                            {formatClaim(translate(`claims.${claim}`), claim)}
                                        </Label>
                                    </li>
                                </TooltipTrigger>
                                <TooltipContent>{translate("Claim", { name: claim })}</TooltipContent>
                            </Tooltip>
                        </TooltipProvider>
                    ))}
                    {availableClaims.map((claim: string) => (
                        <TooltipProvider key={claim}>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <li className="flex items-center gap-2 px-2 py-1">
                                        <Checkbox
                                            id={"claim-" + claim}
                                            checked={claimChecked(claim)}
                                            onCheckedChange={() => handleClaimCheckboxOnChange(claim)}
                                        />
                                        <Label htmlFor={"claim-" + claim}>
                                            {formatClaim(translate(`claims.${claim}`), claim)}
                                        </Label>
                                    </li>
                                </TooltipTrigger>
                                <TooltipContent>{translate("Claim", { name: claim })}</TooltipContent>
                            </Tooltip>
                        </TooltipProvider>
                    ))}
                </ul>
            ) : null}
        </Fragment>
    );
};

export default DecisionFormClaims;
