import React, { Fragment, ReactNode } from "react";

import { Box, Link, Theme, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

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
    const { t: translate } = useTranslation();

    const { classes, cx } = useStyles();

    const registerMessage = props.registered
        ? props.title === "Push Notification"
            ? ""
            : translate("Manage devices")
        : translate("Register device");

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
        <Box id={props.id}>
            <Typography variant={"h6"}>{props.title}</Typography>
            <Box id={"2fa-container"} className={cx(classes.container, stateClass)}>
                <Box className={classes.containerFlex}>{container}</Box>
            </Box>
            {props.onSelectClick && props.registered ? (
                <Link id={"selection-link"} component={"button"} onClick={props.onSelectClick} underline={"hover"}>
                    {translate("Select a Device")}
                </Link>
            ) : null}
            {(props.onRegisterClick && props.title !== "Push Notification") ||
            (props.onRegisterClick && props.title === "Push Notification" && props.duoSelfEnrollment) ? (
                <Link id={"register-link"} component={"button"} onClick={props.onRegisterClick} underline={"hover"}>
                    {registerMessage}
                </Link>
            ) : null}
        </Box>
    );
};

interface NotRegisteredContainerProps {
    title: string;
    duoSelfEnrollment: boolean;
}

function NotRegisteredContainer(props: NotRegisteredContainerProps) {
    const { t: translate } = useTranslation();
    const { classes } = useStyles();

    return (
        <Fragment>
            <Box className={classes.info}>
                <InformationIcon />
            </Box>
            <Typography className={classes.infoTypography}>
                {translate("The resource you're attempting to access requires two-factor authentication")}
            </Typography>
            <Typography className={classes.infoTypography}>
                {props.title === "Push Notification"
                    ? props.duoSelfEnrollment
                        ? translate("Register your first device by clicking on the link below")
                        : translate("Contact your administrator to register a device")
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
    const { classes } = useStyles();

    return (
        <Fragment>
            <Box className={classes.containerMethod}>{props.children}</Box>
            <Typography>{props.explanation}</Typography>
        </Fragment>
    );
}

const useStyles = makeStyles()((theme: Theme) => ({
    container: {
        height: "200px",
    },
    containerFlex: {
        alignContent: "center",
        alignItems: "center",
        display: "flex",
        flexWrap: "wrap",
        height: "100%",
        justifyContent: "center",
        width: "100%",
    },
    containerMethod: {
        marginBottom: theme.spacing(2),
    },
    info: {
        flex: "0 0 100%",
        marginBottom: theme.spacing(2),
    },
    infoTypography: {
        color: "#5858ff",
    },
}));

export default DefaultMethodContainer;
