import React, { Fragment, useCallback, useMemo } from "react";

import { Box, Checkbox, FormControlLabel, List, Theme, Tooltip } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import { formatClaim } from "@services/ConsentOpenIDConnect";

export interface Props {
    onChangeChecked: (claims: string[]) => void;
    claims: null | string[];
    essential_claims: null | string[];
}

const DecisionFormClaims: React.FC<Props> = ({ claims, essential_claims, onChangeChecked }: Props) => {
    const { t: translate } = useTranslation(["consent"]);

    const { classes } = useStyles();

    const checked = useMemo(() => claims || [], [claims]);

    const handleClaimCheckboxOnChange = (event: React.ChangeEvent<HTMLInputElement>) => {
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
                    <Box className={classes.container}>
                        <List className={classes.list}>
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

const useStyles = makeStyles()((theme: Theme) => ({
    container: {
        textAlign: "center",
    },
    list: {
        backgroundColor: theme.palette.background.paper,
        display: "inline-block",
        marginBottom: theme.spacing(2),
        marginTop: theme.spacing(2),
    },
}));

export default DecisionFormClaims;
