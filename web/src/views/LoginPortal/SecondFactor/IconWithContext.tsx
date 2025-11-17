import React, { ReactNode } from "react";

import { Box, Theme } from "@mui/material";
import { makeStyles } from "tss-react/mui";

interface IconWithContextProps {
    icon: ReactNode;
    children: ReactNode;

    className?: string;
}

const IconWithContext = function (props: IconWithContextProps) {
    const { classes, cx } = useStyles({ iconSize: 64 });

    return (
        <Box className={cx(props.className, classes.root)}>
            <Box className={classes.iconContainer}>
                <Box className={classes.icon}>{props.icon}</Box>
            </Box>
            <Box className={classes.context}>{props.children}</Box>
        </Box>
    );
};

const useStyles = makeStyles<{ iconSize: number }>()((theme: Theme, { iconSize }) => ({
    context: {
        display: "block",
    },
    icon: {
        height: iconSize,
        width: iconSize,
    },
    iconContainer: {
        alignItems: "center",
        display: "flex",
        flexDirection: "column",
    },
    root: {},
}));

export default IconWithContext;
