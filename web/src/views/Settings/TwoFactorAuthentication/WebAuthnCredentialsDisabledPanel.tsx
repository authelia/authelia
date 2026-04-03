import { Box, Paper, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

const WebAuthnCredentialsDisabledPanel = function () {
    const { t: translate } = useTranslation("settings");

    return (
        <Paper variant={"outlined"} elevation={24}>
            <Grid container spacing={2} padding={2}>
                <Grid size={{ xs: 12 }}>
                    <Typography variant="h5">{translate("WebAuthn Credentials")}</Typography>
                </Grid>
                <Grid size={{ xs: 12 }} justifyContent={"center"} alignItems={"center"}>
                    <Typography variant={"h6"} color={"secondary"}>
                        {translate(
                            "Your administrator has disabled WebAuthn preventing you from registering WebAuthn Credentials including Passkeys",
                        )}
                        .
                    </Typography>
                </Grid>
                <Grid size={{ xs: 12 }} justifyContent={"center"} alignItems={"center"}>
                    <Typography variant={"body2"}>
                        <Box component={"span"}>
                            {translate(
                                "WebAuthn Credentials are widely considered the most secure means of authentication, regardless of if they're used for Multi-Factor Authentication or Passwordless Authentication",
                            )}
                            .
                        </Box>
                        <Box component={"span"}>
                            {translate(
                                "The decision to disable WebAuthn Credentials when Multi-Factor Authentication is enabled significantly undermines security and is highly inadvisable",
                            )}
                            .
                        </Box>
                    </Typography>
                </Grid>
            </Grid>
        </Paper>
    );
};

export default WebAuthnCredentialsDisabledPanel;
