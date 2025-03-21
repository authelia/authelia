import React from "react";

import { Typography, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import ScaleLoader from "react-spinners/ScaleLoader";

export interface Props {
    message: string;
}

const BaseLoadingPage = function (props: Props) {
    const theme = useTheme();

    return (
        <Grid container sx={{ alignItems: "center", justifyContent: "center", minHeight: "100vh" }}>
            <Grid sx={{ textAlign: "center", display: "inline-block" }}>
                <ScaleLoader width={64} height={64} color={theme.custom.loadingBar} />
                <Typography>{props.message}...</Typography>
            </Grid>
        </Grid>
    );
};

export default BaseLoadingPage;
