import React from "react";

import { Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";

const Authenticated = function () {
    const { t: translate } = useTranslation();

    const styles = useStyles();

    return (
        <div id="authenticated-stage">
            <div className={styles.iconContainer}>
                <SuccessIcon />
            </div>
            <Typography>{translate("Authenticated")}</Typography>
        </div>
    );
};

export default Authenticated;

const useStyles = makeStyles((theme: Theme) => ({
    iconContainer: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
    },
}));
