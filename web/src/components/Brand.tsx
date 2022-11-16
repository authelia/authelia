import React from "react";

import { Grid, Link, Theme } from "@mui/material";
import { grey } from "@mui/material/colors";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

export interface Props {}

const url = "https://www.authelia.com";

const Brand = function (props: Props) {
    const { t: translate } = useTranslation();
    const styles = useStyles();

    return (
        <Grid item xs={12}>
            <Link href={url} target="_blank" underline="hover" className={styles.poweredBy}>
                {translate("Powered by")} Authelia
            </Link>
        </Grid>
    );
};

export default Brand;

const useStyles = makeStyles((theme: Theme) => ({
    poweredBy: {
        fontSize: "0.7em",
        color: grey[500],
    },
}));
