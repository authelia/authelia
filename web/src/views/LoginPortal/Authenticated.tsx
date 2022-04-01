import React, { CSSProperties } from "react";

import { Box, Theme, Typography } from "@mui/material";
import { useTheme } from "@mui/material/styles";
import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";

const Authenticated = function () {
    const theme = useTheme();
    const style = useStyles(theme);

    const { t: translate } = useTranslation("Portal");

    return (
        <Box id="authenticated-stage">
            <Box sx={style.iconContainer}>
                <SuccessIcon />
            </Box>
            <Typography>{translate("Authenticated")}</Typography>
        </Box>
    );
};

export default Authenticated;

const useStyles = (theme: Theme): { [key: string]: CSSProperties } => ({
    iconContainer: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
    },
});
