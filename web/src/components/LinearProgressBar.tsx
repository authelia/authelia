import React from "react";

import { LinearProgress, Theme } from "@mui/material";
import { makeStyles } from "tss-react/mui";

export interface Props {
    value: number;
    height?: number | string;
}

const LinearProgressBar = function (props: Props) {
    const { classes } = useStyles({ props: props });

    return (
        <LinearProgress
            variant="determinate"
            classes={{ determinate: classes.determinate, root: classes.root }}
            value={props.value}
            className={classes.default}
        />
    );
};

const useStyles = makeStyles<{ props: Props }>()((theme: Theme, { props }) => ({
    default: {
        marginTop: theme.spacing(),
    },
    determinate: {
        transition: "transform .2s linear",
    },
    root: {
        height: props.height ? props.height : theme.spacing(),
    },
}));

export default LinearProgressBar;
