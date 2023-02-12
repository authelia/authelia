import React, { ReactNode } from "react";

import { Box, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import classnames from "classnames";

interface IconWithContextProps {
    icon: ReactNode;
    children: ReactNode;

    className?: string;
}

const IconWithContext = function (props: IconWithContextProps) {
    const iconSize = 64;
    const styles = makeStyles((theme: Theme) => ({
        root: {},
        iconContainer: {
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
        },
        icon: {
            width: iconSize,
            height: iconSize,
        },
        context: {
            display: "block",
        },
    }))();

    return (
        <Box className={classnames(props.className, styles.root)}>
            <Box className={styles.iconContainer}>
                <Box className={styles.icon}>{props.icon}</Box>
            </Box>
            <Box className={styles.context}>{props.children}</Box>
        </Box>
    );
};

export default IconWithContext;
