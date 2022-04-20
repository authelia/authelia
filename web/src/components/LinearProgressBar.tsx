import React from "react";

import { LinearProgress } from "@mui/material";
import { Theme } from "@mui/material/styles";
import { CSSProperties, useTheme } from "@mui/styles";

import { StylesProperties } from "@models/StylesProperties";

export interface Props {
    value: number;
    height?: string | number;
    className?: string;
    style?: CSSProperties;
}

const LinearProgressBar = function (props: Props) {
    const theme = useTheme();
    const styles = useStyles(theme, props.height);

    return (
        <LinearProgress
            variant="determinate"
            sx={{
                ...(props.style as CSSProperties),
                ...styles.progressRoot,
                "& .MuiLinearProgress-bar1Determinate": styles.transition,
            }}
            value={props.value}
            className={props.className}
        />
    );
};

export default LinearProgressBar;

const useStyles = (theme: Theme, height?: string | number): StylesProperties => ({
    progressRoot: {
        height: height ? height : theme.spacing(),
    },
    transition: {
        transition: "transform .2s linear",
    },
});
