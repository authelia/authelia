import { Fragment, ReactNode } from "react";

import { Box, Link, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import InformationIcon from "@components/InformationIcon";
import Authenticated from "@views/LoginPortal/Authenticated";

/* eslint-disable no-unused-vars */
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

    let registerMessage;
    if (props.registered) {
        registerMessage = props.title === translate("Push Notification") ? "" : translate("Manage devices");
    } else {
        registerMessage = translate("Register device");
    }

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
            <Box id={"2fa-container"} className={stateClass} sx={{ height: "200px" }}>
                <Box
                    sx={{
                        alignContent: "center",
                        alignItems: "center",
                        display: "flex",
                        flexWrap: "wrap",
                        height: "100%",
                        justifyContent: "center",
                        width: "100%",
                    }}
                >
                    {container}
                </Box>
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
    readonly title: string;
    readonly duoSelfEnrollment: boolean;
}

function NotRegisteredContainer(props: NotRegisteredContainerProps) {
    const { t: translate } = useTranslation();
    let infoText;
    if (props.title === translate("Push Notification")) {
        infoText = props.duoSelfEnrollment
            ? translate("Register your first device by clicking on the link below")
            : translate("Contact your administrator to register a device");
    } else {
        infoText = translate("Register your first device by clicking on the link below");
    }

    return (
        <Fragment>
            <Box sx={{ flex: "0 0 100%", marginBottom: (theme) => theme.spacing(2) }}>
                <InformationIcon />
            </Box>
            <Typography sx={{ color: "#5858ff" }}>
                {translate("The resource you're attempting to access requires two-factor authentication")}
            </Typography>
            <Typography sx={{ color: "#5858ff" }}>{infoText}</Typography>
        </Fragment>
    );
}

interface MethodContainerProps {
    readonly explanation: string;
    readonly children: ReactNode;
}

function MethodContainer(props: MethodContainerProps) {
    return (
        <Fragment>
            <Box sx={{ marginBottom: (theme) => theme.spacing(2) }}>{props.children}</Box>
            <Typography>{props.explanation}</Typography>
        </Fragment>
    );
}

export default DefaultMethodContainer;
