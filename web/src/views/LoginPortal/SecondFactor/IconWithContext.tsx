import React, { CSSProperties, ReactNode } from "react";

import { Box } from "@mui/material";

interface IconWithContextProps {
    icon: ReactNode;
    children: ReactNode;

    className?: string;
    styles?: { [key: string]: CSSProperties };
}

const IconWithContext = function (props: IconWithContextProps) {
    const iconSize = 64;
    const styles: { [key: string]: CSSProperties } = {
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
    };

    return (
        <Box sx={{ ...styles.root, ...props.styles }} className={props.className}>
            <Box sx={styles.iconContainer}>
                <Box sx={styles.icon}>{props.icon}</Box>
            </Box>
            <Box sx={styles.context}>{props.children}</Box>
        </Box>
    );
};

export default IconWithContext;
