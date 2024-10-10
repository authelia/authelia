import React from "react";

import { Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import FailureIcon from "@components/FailureIcon";

const GenericError = function () {
    const styles = useStyles();
    const { t: translate } = useTranslation();
    return (
        <div id="generic-error-stage">
            <div className={styles.iconContainer}>
                <FailureIcon />
            </div>
            <Typography>{translate("Error")}</Typography>
        </div>
    );
};

export default GenericError;

const useStyles = makeStyles((theme: Theme) => ({
    iconContainer: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
    },
}));
