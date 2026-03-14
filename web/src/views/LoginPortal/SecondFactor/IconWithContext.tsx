import { ReactNode } from "react";

import { Box } from "@mui/material";

interface IconWithContextProps {
    icon: ReactNode;
    children: ReactNode;

    className?: string;
}

const IconWithContext = function (props: IconWithContextProps) {
    return (
        <Box className={props.className}>
            <Box sx={{ alignItems: "center", display: "flex", flexDirection: "column" }}>
                <Box sx={{ height: 64, width: 64 }}>{props.icon}</Box>
            </Box>
            <Box sx={{ display: "block" }}>{props.children}</Box>
        </Box>
    );
};

export default IconWithContext;
