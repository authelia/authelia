import { ChangeEvent, FC, Fragment, useCallback, useMemo } from "react";

import { Box, Checkbox, FormControlLabel, List, Tooltip } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import { formatClaim } from "@services/ConsentOpenIDConnect";

export interface Props {
    onChangeChecked: (_claims: string[]) => void;
    claims: null | string[];
    essential_claims: null | string[];
}

const DecisionFormClaims: FC<Props> = ({ claims, essential_claims, onChangeChecked }: Props) => {
    const { t: translate } = useTranslation(["consent"]);

    const checked = useMemo(() => claims || [], [claims]);

    const handleClaimCheckboxOnChange = (event: ChangeEvent<HTMLInputElement>) => {
        const checking = !checked.includes(event.target.value);

        if (checking) {
            onChangeChecked([...checked, event.target.value]);
        } else {
            onChangeChecked(checked.filter((value) => value !== event.target.value));
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
                <Grid size={{ xs: 12 }}>
                    <Box sx={{ textAlign: "center" }}>
                        <List
                            sx={{
                                backgroundColor: (theme) => theme.palette.background.paper,
                                display: "inline-block",
                                marginBottom: (theme) => theme.spacing(2),
                                marginTop: (theme) => theme.spacing(2),
                            }}
                        >
                            {essential_claims?.map((claim: string) => (
                                <Tooltip key={`${claim}-essential`} title={translate("Claim", { name: claim })}>
                                    <FormControlLabel
                                        control={<Checkbox id={`claim-${claim}-essential`} disabled checked />}
                                        label={formatClaim(translate(`claims.${claim}`), claim)}
                                    />
                                </Tooltip>
                            ))}
                            {claims?.map((claim: string) => (
                                <Tooltip key={claim} title={translate("Claim", { name: claim })}>
                                    <FormControlLabel
                                        control={
                                            <Checkbox
                                                id={"claim-" + claim}
                                                value={claim}
                                                checked={claimChecked(claim)}
                                                onChange={handleClaimCheckboxOnChange}
                                            />
                                        }
                                        label={formatClaim(translate(`claims.${claim}`), claim)}
                                    />
                                </Tooltip>
                            ))}
                        </List>
                    </Box>
                </Grid>
            ) : null}
        </Fragment>
    );
};

export default DecisionFormClaims;
