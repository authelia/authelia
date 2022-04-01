import React, { CSSProperties, ReactNode } from "react";

import { Box } from "@mui/material";

interface Props {
    icon: ReactNode;
    children: ReactNode;

    className?: string;
    styles?: { [key: string]: CSSProperties };
}

const IconWithContext = function (props: Props) {
    const style = useStyles();

    return (
        <Box sx={{ ...style.root, ...props.styles }} className={props.className}>
            <Box sx={style.iconContainer}>
                <Box sx={style.icon}>{props.icon}</Box>
            </Box>
            <Box sx={style.context}>{props.children}</Box>
        </Box>
    );
};

export default IconWithContext;

const useStyles = (): { [key: string]: CSSProperties } => ({
    root: {},
    iconContainer: {
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
    },
    icon: {
        width: 64,
        height: 64,
    },
    context: {
        display: "block",
    },
});
