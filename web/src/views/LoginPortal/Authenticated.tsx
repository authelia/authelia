import React from "react";

import { Typography, makeStyles } from "@material-ui/core";
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
