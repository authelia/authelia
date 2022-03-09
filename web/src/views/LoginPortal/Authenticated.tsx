import React from "react";

import { Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";

const Authenticated = function () {
    const classes = useStyles();
    const { t: translate } = useTranslation("Portal");
    return (
        <div id="authenticated-stage">
            <div className={classes.iconContainer}>
                <SuccessIcon />
            </div>
            <Typography>{translate("Authenticated")}</Typography>
        </div>
    );
};

export default Authenticated;

const useStyles = makeStyles((theme) => ({
    iconContainer: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
    },
}));
