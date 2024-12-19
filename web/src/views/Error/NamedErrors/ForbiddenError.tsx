import React from "react";

import { Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";

import FailureIcon from "@components/FailureIcon";
import { RedirectionURL } from "@constants/SearchParams";

const ForbiddenError = function () {
    const styles = useStyles();
    const { t: translate } = useTranslation();
    const [searchParams] = useSearchParams();

    const getMessage = () => {
        if (!searchParams.has(RedirectionURL)) {
            return "Access Denied";
        } else {
            return "Access Denied To:";
        }
    };

    return (
        <div id="access-denied-stage">
            <div className={styles.iconContainer}>
                <FailureIcon />
            </div>
            <Typography>{translate(getMessage())}</Typography>
            <Typography className={styles.textEllipsis}>{searchParams.get(RedirectionURL) || ""}</Typography>
        </div>
    );
};

export default ForbiddenError;

const useStyles = makeStyles((theme: Theme) => ({
    iconContainer: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
    },
    textEllipsis: {
        overflow: "hidden",
        textOverflow: "ellipsis",
    },
}));
