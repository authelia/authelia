import React from "react";

import { LinearProgress, useTheme } from "@mui/material";

export interface Props {
    value: number;
    height?: string | number;
}

const LinearProgressBar = function (props: Props) {
    const theme = useTheme();

    return (
        <LinearProgress
            variant="determinate"
            sx={{
                marginTop: theme.spacing(),
                "& .MuiLinearProgress-root": {
                    height: props.height ? props.height : theme.spacing(),
                },
                "& .MuiLinearProgress-bar1Determinate": {
                    transition: "transform .2s linear",
                },
            }}
            value={props.value}
        />
    );
};

export default LinearProgressBar;
