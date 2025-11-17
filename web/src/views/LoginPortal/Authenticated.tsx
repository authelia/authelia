import React from "react";

import { Box, Theme, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import SuccessIcon from "@components/SuccessIcon";

const Authenticated = function () {
    const { t: translate } = useTranslation();

    const { classes } = useStyles();

    return (
        <Box id="authenticated-stage">
            <Box className={classes.iconContainer}>
                <SuccessIcon />
            </Box>
            <Typography>{translate("Authenticated")}</Typography>
        </Box>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    iconContainer: {
        flex: "0 0 100%",
        marginBottom: theme.spacing(2),
    },
}));

export default Authenticated;
