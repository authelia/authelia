import React, { Fragment } from "react";

import { Divider, Link, Theme } from "@mui/material";
import { grey } from "@mui/material/colors";
import Grid from "@mui/material/Unstable_Grid2/Grid2";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import PrivacyPolicyLink from "@components/PrivacyPolicyLink";
import { getPrivacyPolicyEnabled } from "@utils/Configuration";

export interface Props {}

const url = "https://www.authelia.com";

const Brand = function (props: Props) {
    const { t: translate } = useTranslation();

    const styles = useStyles();
    const privacyEnabled = getPrivacyPolicyEnabled();

    return (
        <Grid container xs={12} alignItems="center" justifyContent="center">
            <Grid xs={4}>
                <Link href={url} target="_blank" underline="hover" className={styles.links}>
                    {translate("Powered by")} Authelia
                </Link>
            </Grid>
            {privacyEnabled ? (
                <Fragment>
                    <Divider orientation="vertical" flexItem variant="middle" />
                    <Grid xs={4}>
                        <PrivacyPolicyLink className={styles.links} />
                    </Grid>
                </Fragment>
            ) : null}
        </Grid>
    );
};

const useStyles = makeStyles((theme: Theme) => ({
    links: {
        fontSize: "0.7rem",
        color: grey[500],
    },
}));

export default Brand;
