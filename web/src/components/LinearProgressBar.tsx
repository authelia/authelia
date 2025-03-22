import React from "react";

import { LinearProgress, Theme } from "@mui/material";
import { makeStyles } from "tss-react/mui";

export interface Props {
    value: number;
    height?: string | number;
}

const LinearProgressBar = function (props: Props) {
    const { classes } = useStyles({ props: props });

    return (
        <LinearProgress
            variant="determinate"
            classes={{ root: classes.root, determinate: classes.determinate }}
            value={props.value}
            className={classes.default}
        />
    );
};

const useStyles = makeStyles<{ props: Props }>()((theme: Theme, { props }) => ({
    root: {
        height: props.height ? props.height : theme.spacing(),
    },
    determinate: {
        transition: "transform .2s linear",
    },
    default: {
        marginTop: theme.spacing(),
    },
}));

export default LinearProgressBar;
