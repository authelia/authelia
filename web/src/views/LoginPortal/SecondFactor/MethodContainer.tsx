import React, { ReactNode, Fragment } from "react";
import { makeStyles, Typography, Link } from "@material-ui/core";

interface MethodContainerProps {
    title: string;
    explanation: string;
    children: ReactNode;

    onRegisterClick?: () => void;
}

export default function (props: MethodContainerProps) {
    const style = useStyles();
    return (
        <Fragment>
            <Typography variant="h6">{props.title}</Typography>
            <div className={style.icon}>{props.children}</div>
            <Typography>{props.explanation}</Typography>
            {props.onRegisterClick
                ? <Link component="button" onClick={props.onRegisterClick}>
                    Not registered yet?
                </Link>
                : null}
        </Fragment>
    )
}

const useStyles = makeStyles(theme => ({
    icon: {
        paddingTop: theme.spacing(2),
        paddingBottom: theme.spacing(2),
    },
}));