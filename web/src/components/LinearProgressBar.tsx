import React from "react";
import { makeStyles, LinearProgress } from "@material-ui/core";
import { CSSProperties } from "@material-ui/styles";

export interface Props {
    value: number;
    height?: number;
    className?: string;
    style?: CSSProperties;
}

export default function (props: Props) {
    const style = makeStyles(theme => ({
        progressRoot: {
            height: props.height ? props.height : theme.spacing(),
        },
        transition: {
            transition: "transform .2s linear",
        }
    }))();
    return (
        <LinearProgress
            style={props.style}
            variant="determinate"
            classes={{
                root: style.progressRoot,
                bar1Determinate: style.transition
            }}
            value={props.value}
            className={props.className} />
    )
}
