import React from "react";
import SuccessIcon from "../../components/SuccessIcon";
import { Typography, makeStyles } from "@material-ui/core";

export default function () {
    const classes = useStyles();
    return (
        <div id="authenticated-stage">
            <div className={classes.iconContainer}>
                <SuccessIcon />
            </div>
            <Typography>Authenticated</Typography>
        </div>
    )
}

const useStyles = makeStyles(theme => ({
    iconContainer: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%"
    }
}))