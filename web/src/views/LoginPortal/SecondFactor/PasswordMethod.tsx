import { useTranslation } from "react-i18next";

import { AuthenticationLevel } from "@services/State";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";
import PasswordForm from "@views/LoginPortal/SecondFactor/PasswordForm";

export interface Props {
    id: string;

    authenticationLevel: AuthenticationLevel;

    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
}

const PasswordMethod = function (props: Props) {
    const { t: translate } = useTranslation();

    const methodState =
        props.authenticationLevel === AuthenticationLevel.TwoFactor
            ? MethodContainerState.ALREADY_AUTHENTICATED
            : MethodContainerState.METHOD;

    return (
        <MethodContainer
            id={props.id}
            title={translate("Password")}
            explanation={translate("Enter your password to confirm your identity")}
            duoSelfEnrollment={false}
            registered={true}
            state={methodState}
        >
            <PasswordForm onAuthenticationSuccess={props.onAuthenticationSuccess} />
        </MethodContainer>
    );
};

export default PasswordMethod;
