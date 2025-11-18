import { Box, Typography, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import ScaleLoader from "react-spinners/ScaleLoader";
import { makeStyles } from "tss-react/mui";

export interface Props {
    message: string;
}

const BaseLoadingPage = function (props: Props) {
    const theme = useTheme();
    const { classes } = useStyles();

    return (
        <Grid container className={classes.gridOuter}>
            <Grid className={classes.gridInner}>
                <Box padding={theme.spacing(2)}>
                    <ScaleLoader color={theme.custom.loadingBar} speedMultiplier={1.5} />
                </Box>
                <Box padding={theme.spacing(2)}>
                    <Typography>{props.message}...</Typography>
                </Box>
            </Grid>
        </Grid>
    );
};

const useStyles = makeStyles()(() => ({
    gridInner: {
        display: "inline-block",
        textAlign: "center",
    },
    gridOuter: {
        alignItems: "center",
        justifyContent: "center",
        minHeight: "100vh",
    },
}));

export default BaseLoadingPage;
