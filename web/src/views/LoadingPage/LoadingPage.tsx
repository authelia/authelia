import React from "react";

import { Typography, Grid } from "@material-ui/core";
import ReactLoading from "react-loading";

const LoadingPage = function () {
    return (
        <Grid container alignItems="center" justify="center" style={{ minHeight: "100vh" }}>
            <Grid item style={{ textAlign: "center", display: "inline-block" }}>
                <ReactLoading width={64} height={64} color="black" type="bars" />
                <Typography>Loading...</Typography>
            </Grid>
        </Grid>
    );
};

export default LoadingPage;
