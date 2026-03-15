import { FC, Fragment, useCallback, useMemo } from "react";

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

    const hasClaims = essential_claims || claims;

    return (
        <Fragment>
            {hasClaims ? (
                <div className="w-full">
                    <div className="text-center">
                        <ul className="my-4 inline-block list-none bg-card">
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
                            {claims?.map((claim: string) => (
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
                    </div>
                </div>
            ) : null}
        </Fragment>
    );
};

export default DecisionFormClaims;
