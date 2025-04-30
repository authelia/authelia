import React, { Fragment, useCallback, useEffect, useState } from "react";

import { Box, Checkbox, FormControlLabel, List, Theme, Tooltip } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import { formatClaim } from "@services/ConsentOpenIDConnect";

export interface Props {
    onChangeChecked: (claims: string[]) => void;
    claims: string[] | null;
    essential_claims: string[] | null;
}

const OpenIDConnectConsentDecisionFormClaims: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["portal", "consent"]);

    const { classes } = useStyles();

    const [checked, setChecked] = useState<string[]>([]);

    const handleClaimCheckboxOnChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setChecked((prevState) => {
            const checking = !prevState.includes(event.target.value);

            if (checking) {
                return [...prevState, event.target.value];
            } else {
                return prevState.filter((value) => value !== event.target.value);
            }
        });
    };

    useEffect(() => {
        props.onChangeChecked(checked);
    }, [checked, props]);

    const claimChecked = useCallback(
        (claim: string) => {
            return checked.includes(claim);
        },
        [checked],
    );

    const hasClaims = props?.essential_claims || props?.claims;

    return (
        <Fragment>
            {hasClaims ? (
                <Grid size={{ xs: 12 }}>
                    <Box className={classes.container}>
                        <List className={classes.list}>
                            {props.essential_claims?.map((claim: string) => (
                                <Tooltip title={translate("Claim", { name: claim, ns: "consent" })}>
                                    <FormControlLabel
                                        control={<Checkbox id={`claim-${claim}-essential`} disabled checked />}
                                        label={formatClaim(translate(`claims.${claim}`, { ns: "consent" }), claim)}
                                    />
                                </Tooltip>
                            ))}
                            {props.claims?.map((claim: string) => (
                                <Tooltip title={translate("Claim", { name: claim, ns: "consent" })}>
                                    <FormControlLabel
                                        control={
                                            <Checkbox
                                                id={"claim-" + claim}
                                                value={claim}
                                                checked={claimChecked(claim)}
                                                onChange={handleClaimCheckboxOnChange}
                                            />
                                        }
                                        label={formatClaim(translate(`claims.${claim}`, { ns: "consent" }), claim)}
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
        display: "inline-block",
        backgroundColor: theme.palette.background.paper,
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));

export default OpenIDConnectConsentDecisionFormClaims;
