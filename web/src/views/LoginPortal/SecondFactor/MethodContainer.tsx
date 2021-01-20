import React, { ReactNode, Fragment } from "react";

import { makeStyles, Typography, Link, useTheme } from "@material-ui/core";
import classnames from "classnames";

import InformationIcon from "../../../components/InformationIcon";
import Authenticated from "../Authenticated";

export enum State {
    ALREADY_AUTHENTICATED = 1,
    NOT_REGISTERED = 2,
    METHOD = 3,
}

export interface Props {
    id: string;
    title: string;
    explanation: string;
    state: State;
    children: ReactNode;

    onRegisterClick?: () => void;
}

const DefaultMethodContainer = function (props: Props) {
    const style = useStyles();

    let container: ReactNode;
    let stateClass: string = "";
    switch (props.state) {
        case State.ALREADY_AUTHENTICATED:
            container = <Authenticated />;
            stateClass = "state-already-authenticated";
            break;
        case State.NOT_REGISTERED:
            container = <NotRegisteredContainer />;
            stateClass = "state-not-registered";
            break;
        case State.METHOD:
            container = <MethodContainer explanation={props.explanation}>{props.children}</MethodContainer>;
            stateClass = "state-method";
            break;
    }

    return (
        <div id={props.id}>
            <Typography variant="h6">{props.title}</Typography>
            <div className={classnames(style.container, stateClass)} id="2fa-container">
                <div className={style.containerFlex}>{container}</div>
            </div>
            {props.onRegisterClick ? (
                <Link component="button" id="register-link" onClick={props.onRegisterClick}>
                    Not registered yet?
                </Link>
            ) : null}
        </div>
    );
};

export default DefaultMethodContainer;

const useStyles = makeStyles((theme) => ({
    container: {
        height: "200px",
    },
    containerFlex: {
        display: "flex",
        flexWrap: "wrap",
        height: "100%",
        width: "100%",
        alignItems: "center",
        alignContent: "center",
        justifyContent: "center",
    },
}));

function NotRegisteredContainer() {
    const theme = useTheme();
    return (
        <Fragment>
            <div style={{ marginBottom: theme.spacing(2), flex: "0 0 100%" }}>
                <InformationIcon />
            </div>
            <Typography style={{ color: "#5858ff" }}>
                Register your first device by clicking on the link below
            </Typography>
        </Fragment>
    );
}

interface MethodContainerProps {
    explanation: string;
    children: ReactNode;
}

function MethodContainer(props: MethodContainerProps) {
    const theme = useTheme();
    return (
        <Fragment>
            <div style={{ marginBottom: theme.spacing(2) }}>{props.children}</div>
            <Typography>{props.explanation}</Typography>
        </Fragment>
    );
}
