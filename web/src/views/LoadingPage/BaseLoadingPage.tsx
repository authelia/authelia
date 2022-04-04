import React from "react";

import { Grid, Typography, useTheme } from "@material-ui/core";
import ReactLoading from "react-loading";

export interface Props {
    message: string;
}

const BaseLoadingPage = function (props: Props) {
    const theme = useTheme();
    return (
        <Grid container alignItems="center" justifyContent="center" style={{ minHeight: "100vh" }}>
            <Grid item style={{ textAlign: "center", display: "inline-block" }}>
                <ReactLoading width={64} height={64} color={theme.custom.loadingBar} type="bars" />
                <Typography>{props.message}...</Typography>
            </Grid>
        </Grid>
    );
};

export default BaseLoadingPage;
