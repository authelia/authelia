import React from "react";

import { LinearProgress, Theme, useTheme } from "@mui/material";
import { CSSProperties } from "@mui/styles";

export interface Props {
    value: number;
    height?: number;
    className?: string;
    style?: CSSProperties;
}

const LinearProgressBar = function (props: Props) {
    const theme = useTheme();
    const style = useStyles(theme, props);

    return (
        /*
        TODO: Check this component.
               classes={{
                root: style.progressRoot,
                bar1Determinate: style.transition,
            }}
         */
        <LinearProgress
            style={props.style as React.CSSProperties}
            variant="determinate"
            sx={{ ...style.progressRoot, ...style.transition }}
            value={props.value}
            className={props.className}
        />
    );
};

export default LinearProgressBar;

const useStyles = (theme: Theme, props: Props): { [key: string]: CSSProperties } => ({
    progressRoot: {
        height: props.height ? props.height : theme.spacing(),
    },
    transition: {
        transition: "transform .2s linear",
    },
});
