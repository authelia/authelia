import { Box, Typography, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import { ScaleLoader } from "react-spinners";

export interface Props {
    message: string;
}

const BaseLoadingPage = function (props: Props) {
    const theme = useTheme();

    return (
        <Grid container sx={{ alignItems: "center", justifyContent: "center", minHeight: "100vh" }}>
            <Grid sx={{ display: "inline-block", textAlign: "center" }}>
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

export default BaseLoadingPage;
