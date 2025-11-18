import { Box, Paper, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

const SettingsView = function () {
    const { t: translate } = useTranslation("settings");

    return (
        <Paper variant={"outlined"}>
            <Box sx={{ p: 3 }}>
                <Typography variant={"h4"} textAlign={"center"} mb={1}>
                    {translate("User Settings")}
                </Typography>
                <Typography textAlign={"center"} my={1}>
                    {translate(
                        "This is the user settings area at the present time it's very minimal but will include new features in the near future",
                    )}
                </Typography>
                <Typography textAlign={"center"} my={1}>
                    {translate("To view the currently available options select the menu icon at the top left")}
                </Typography>
            </Box>
        </Paper>
    );
};

export default SettingsView;
