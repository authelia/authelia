import React from "react";

import { Grid, Theme, Typography, useTheme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import ReactLoading from "react-loading";

export interface Props {
    message: string;
}

const BaseLoadingPage = function (props: Props) {
    const theme = useTheme();
    const styles = useStyles();

    return (
        <Grid container className={styles.gridOuter}>
            <Grid item className={styles.gridInner}>
                <ReactLoading width={64} height={64} color={theme.custom.loadingBar} type="bars" />
                <Typography>{props.message}...</Typography>
            </Grid>
        </Grid>
    );
};

const useStyles = makeStyles((theme: Theme) => ({
    gridOuter: {
        alignItems: "center",
        justifyContent: "center",
        minHeight: "100vh",
    },
    gridInner: {
        textAlign: "center",
        display: "inline-block",
    },
}));

export default BaseLoadingPage;
