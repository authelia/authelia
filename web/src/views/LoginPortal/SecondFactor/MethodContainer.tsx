import { Fragment, ReactNode } from "react";

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
        <div id={props.id}>
            <h6 className="text-xl font-medium">{props.title}</h6>
            <div id={"2fa-container"} className={`${stateClass} h-[200px]`}>
                <div className="flex h-full w-full flex-wrap content-center items-center justify-center">
                    {container}
                </div>
            </div>
            {props.onSelectClick && props.registered ? (
                <button
                    id={"selection-link"}
                    className="text-base text-primary underline-offset-4 hover:underline"
                    onClick={props.onSelectClick}
                >
                    {translate("Select a Device")}
                </button>
            ) : null}
            {(props.onRegisterClick && props.title !== "Push Notification") ||
            (props.onRegisterClick && props.title === "Push Notification" && props.duoSelfEnrollment) ? (
                <button
                    id={"register-link"}
                    className="text-base text-primary underline-offset-4 hover:underline"
                    onClick={props.onRegisterClick}
                >
                    {registerMessage}
                </button>
            ) : null}
        </div>
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
            <div className="mb-4 flex-[0_0_100%]">
                <InformationIcon />
            </div>
            <p className="text-[#5858ff]">
                {translate("The resource you're attempting to access requires two-factor authentication")}
            </p>
            <p className="text-[#5858ff]">{infoText}</p>
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
            <div className="mb-4">{props.children}</div>
            <p>{props.explanation}</p>
        </Fragment>
    );
}

export default DefaultMethodContainer;
