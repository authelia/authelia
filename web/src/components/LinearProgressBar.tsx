import React from "react";

import { LinearProgress, Theme } from "@mui/material";
import { CSSProperties } from "@mui/styles";
import makeStyles from "@mui/styles/makeStyles";

export interface Props {
    value: number;
    height?: string | number;
    className?: string;
    style?: CSSProperties;
}

const LinearProgressBar = function (props: Props) {
    const styles = makeStyles((theme: Theme) => ({
        progressRoot: {
            height: props.height ? props.height : theme.spacing(),
        },
        transition: {
            transition: "transform .2s linear",
        },
    }))();

    return (
        <LinearProgress
            style={props.style as React.CSSProperties}
            variant="determinate"
            classes={{
                root: styles.progressRoot,
                bar1Determinate: styles.transition,
            }}
            value={props.value}
            className={props.className}
        />
    );
};

export default LinearProgressBar;
