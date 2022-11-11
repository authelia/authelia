import React from "react";

import { Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { Navigate } from "react-router-dom";

import SuccessIcon from "@components/SuccessIcon";
import { IndexRoute } from "@constants/Routes";

const Authenticated = function () {
    const styles = useStyles();
    const { t: translate } = useTranslation();

    return (
        <div id="authenticated-stage">
            <div className={styles.iconContainer}>
                <SuccessIcon />
            </div>
            <Typography>{translate("Authenticated")}</Typography>
            <Navigate to={IndexRoute} />;
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
