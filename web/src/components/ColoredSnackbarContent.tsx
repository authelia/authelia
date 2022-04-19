import React from "react";

import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import ErrorIcon from "@mui/icons-material/Error";
import InfoIcon from "@mui/icons-material/Info";
import WarningIcon from "@mui/icons-material/Warning";
import { SnackbarContent } from "@mui/material";
import { amber, green } from "@mui/material/colors";
import { SnackbarContentProps } from "@mui/material/SnackbarContent";
import makeStyles from "@mui/styles/makeStyles";
import classnames from "classnames";

const variantIcon = {
    success: CheckCircleIcon,
    warning: WarningIcon,
    error: ErrorIcon,
    info: InfoIcon,
};

export type Level = keyof typeof variantIcon;

export interface Props extends SnackbarContentProps {
    className?: string;
    level: Level;
    message: string;
}

const ColoredSnackbarContent = function (props: Props) {
    const classes = useStyles();
    const Icon = variantIcon[props.level];

    const { className, variant, message, ...others } = props;

    return (
        <SnackbarContent
            className={classnames(classes[props.level], className)}
            message={
                <span className={classes.message}>
                    <Icon className={classnames(classes.icon, classes.iconVariant)} />
                    {message}
                </span>
            }
            {...others}
        />
    );
};

export default ColoredSnackbarContent;

const useStyles = makeStyles((theme) => ({
    success: {
        backgroundColor: green[600],
    },
    error: {
        backgroundColor: theme.palette.error.dark,
    },
    info: {
        backgroundColor: theme.palette.primary.main,
    },
    warning: {
        backgroundColor: amber[700],
    },
    icon: {
        fontSize: 20,
    },
    iconVariant: {
        opacity: 0.9,
        marginRight: theme.spacing(1),
    },
    message: {
        display: "flex",
        alignItems: "center",
    },
}));
