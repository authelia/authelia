import React, { CSSProperties } from "react";

import { Box, Typography } from "@mui/material";
import { useTheme } from "@mui/material/styles";
import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";

const Authenticated = function () {
    const theme = useTheme();

    const styles: { [key: string]: CSSProperties } = {
        iconContainer: {
            marginBottom: theme.spacing(2),
            flex: "0 0 100%",
        },
    };

    const { t: translate } = useTranslation("Portal");

    return (
        <Box id="authenticated-stage">
            <Box sx={styles.iconContainer}>
                <SuccessIcon />
            </Box>
            <Typography>{translate("Authenticated")}</Typography>
        </Box>
    );
};

export default Authenticated;
