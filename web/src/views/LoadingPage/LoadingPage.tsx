import React from "react";
import ReactLoading from "react-loading";
import { Typography, Grid } from "@material-ui/core";
import { useTheme } from '../../hooks/Theme';

var color = "#000";
const theme = useTheme();
if (theme == "dark") {
  color = "#929aa5"
} else {
  color = "#000"
}

export default function () {
    return (
        <Grid container alignItems="center" justify="center" style={{ minHeight: "100vh" }}>
            <Grid item style={{ textAlign: "center", display: "inline-block" }}>
                <ReactLoading width={64} height={64} color={color} type="bars" />
                <Typography>Loading...</Typography>
            </Grid>
        </Grid>
    );
}
