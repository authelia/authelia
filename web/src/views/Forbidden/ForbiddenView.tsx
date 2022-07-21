import React from "react";

import { Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import { useRedirectionURL } from "@hooks/RedirectionURL";
import LoginLayout from "@layouts/LoginLayout";

export interface Props {}

const ForbiddenView = function (props: Props) {
    const styles = useStyles();
    const redirectionURL = useRedirectionURL();
    const { t: translate } = useTranslation();

    return (
        <LoginLayout title={translate("Forbidden")}>
            <Typography className={styles.typo}>
                {translate("You're forbidden to access the following URL:")}
            </Typography>
            <Typography className={styles.typo}>{redirectionURL}</Typography>
        </LoginLayout>
    );
};

export default ForbiddenView;

const useStyles = makeStyles((theme: Theme) => ({
    typo: {
        padding: theme.spacing(),
    },
}));
