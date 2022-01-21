import React from "react";

import { useTheme, Typography, Grid } from "@material-ui/core";
import { useTranslation } from "react-i18next";
import ReactLoading from "react-loading";

const LoadingPage = function () {
    const theme = useTheme();
    const { t: translate } = useTranslation("Portal");
    return (
        <Grid container alignItems="center" justifyContent="center" style={{ minHeight: "100vh" }}>
            <Grid item style={{ textAlign: "center", display: "inline-block" }}>
                <ReactLoading width={64} height={64} color={theme.custom.loadingBar} type="bars" />
                <Typography>{translate("Loading")}...</Typography>
            </Grid>
        </Grid>
    );
};

export default LoadingPage;
