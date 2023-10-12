import React from "react";

import { Box, Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";

const Authenticated = function () {
    const { t: translate } = useTranslation();

    const styles = useStyles();

    return (
        <Box id="authenticated-stage">
            <Box className={styles.iconContainer}>
                <SuccessIcon />
            </Box>
            <Typography>{translate("Authenticated")}</Typography>
        </Box>
    );
};

const useStyles = makeStyles((theme: Theme) => ({
    iconContainer: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
    },
}));

export default Authenticated;
