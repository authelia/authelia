import React, { CSSProperties } from "react";

import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import ErrorIcon from "@mui/icons-material/Error";
import InfoIcon from "@mui/icons-material/Info";
import WarningIcon from "@mui/icons-material/Warning";
import { Box, SnackbarContent, Theme, useTheme } from "@mui/material";
import { amber, green } from "@mui/material/colors";
import { SnackbarContentProps } from "@mui/material/SnackbarContent";

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
    const theme = useTheme();
    const style = useStyles(theme);

    const Icon = variantIcon[props.level];

    const { className, variant, message, ...others } = props;

    let sx = style[props.level];
    if (className !== undefined) {
        sx = { ...sx, ...style[className] };
    }

    return (
        <SnackbarContent
            sx={sx}
            message={
                <Box component="span" sx={style.message}>
                    <Icon sx={{ ...style.icon, ...style.iconVariant }} />
                    {message}
                </Box>
            }
            {...others}
        />
    );
};

export default ColoredSnackbarContent;

const useStyles = (theme: Theme): { [key: string]: CSSProperties } => ({
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
});
