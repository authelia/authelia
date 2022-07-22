import React from "react";

import { Link, Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import { useErrorValues } from "@hooks/ErrorValues";
import LoginLayout from "@layouts/LoginLayout";

export interface Props {}

const ErrorView = function (props: Props) {
    const styles = useStyles();
    const values = useErrorValues();
    const { t: translate } = useTranslation();

    return (
        <LoginLayout title={values.title ? translate(values.title) : "Error"}>
            {values.code ? (
                <Typography className={styles.typo} variant={"h3"}>
                    {values.code} {values.message ? values.message : ""}
                </Typography>
            ) : null}
            {values.description ? (
                <Typography className={styles.typo}>{translate(values.description)}.</Typography>
            ) : null}
            {values.url ? (
                <Typography className={styles.typo} variant={"subtitle1"}>
                    This error is in relation to the following URL: <Link href={values.url}>{values.url}</Link>
                </Typography>
            ) : null}
        </LoginLayout>
    );
};

export default ErrorView;

const useStyles = makeStyles((theme: Theme) => ({
    typo: {
        padding: theme.spacing(),
    },
}));
