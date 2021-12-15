import React, { ReactNode, Fragment } from "react";

import { makeStyles, Typography, Link, useTheme } from "@material-ui/core";
import classnames from "classnames";
import { useTranslation } from "react-i18next";

import InformationIcon from "@components/InformationIcon";
import Authenticated from "@views/LoginPortal/Authenticated";

export enum State {
    ALREADY_AUTHENTICATED = 1,
    NOT_REGISTERED = 2,
    METHOD = 3,
}

export interface Props {
    id: string;
    title: string;
    duoSelfEnrollment: boolean;
    registered: boolean;
    explanation: string;
    state: State;
    children: ReactNode;

    onRegisterClick?: () => void;
    onSelectClick?: () => void;
}

const DefaultMethodContainer = function (props: Props) {
    const style = useStyles();
    const { t: translate } = useTranslation("Portal");
    const registerMessage = props.registered
        ? props.title === "Push Notification"
            ? ""
            : translate("Lost your device?")
        : translate("Register device");
    const selectMessage = translate("Select a Device");

    let container: ReactNode;
    let stateClass: string = "";
    switch (props.state) {
        case State.ALREADY_AUTHENTICATED:
            container = <Authenticated />;
            stateClass = "state-already-authenticated";
            break;
        case State.NOT_REGISTERED:
            container = <NotRegisteredContainer title={props.title} duoSelfEnrollment={props.duoSelfEnrollment} />;
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
            {props.onSelectClick && props.registered ? (
                <Link component="button" id="selection-link" onClick={props.onSelectClick}>
                    {selectMessage}
                </Link>
            ) : null}
            {(props.onRegisterClick && props.title !== "Push Notification") ||
            (props.onRegisterClick && props.title === "Push Notification" && props.duoSelfEnrollment) ? (
                <Link component="button" id="register-link" onClick={props.onRegisterClick}>
                    {registerMessage}
                </Link>
            ) : null}
        </div>
    );
};

export default DefaultMethodContainer;

const useStyles = makeStyles(() => ({
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

interface NotRegisteredContainerProps {
    title: string;
    duoSelfEnrollment: boolean;
}

function NotRegisteredContainer(props: NotRegisteredContainerProps) {
    const { t: translate } = useTranslation("Portal");
    const theme = useTheme();
    return (
        <Fragment>
            <div style={{ marginBottom: theme.spacing(2), flex: "0 0 100%" }}>
                <InformationIcon />
            </div>
            <Typography style={{ color: "#5858ff" }}>
                {translate("The resource you're attempting to access requires two-factor authentication")}
            </Typography>
            <Typography style={{ color: "#5858ff" }}>
                {props.title === "Push Notification"
                    ? props.duoSelfEnrollment
                        ? translate("Register your first device by clicking on the link below")
                        : translate("Contact your administrator to register a device.")
                    : translate("Register your first device by clicking on the link below")}
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
