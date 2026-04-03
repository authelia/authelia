import { AuthenticationLevel } from "@services/State";
import MethodContainer, { State as MethodContainerState } from "@views/LoginPortal/SecondFactor/MethodContainer";
import PasswordForm from "@views/LoginPortal/SecondFactor/PasswordForm";

export interface Props {
    id: string;

    authenticationLevel: AuthenticationLevel;

    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
}

const PasswordMethod = function (props: Props) {
    const methodState =
        props.authenticationLevel === AuthenticationLevel.TwoFactor
            ? MethodContainerState.ALREADY_AUTHENTICATED
            : MethodContainerState.METHOD;

    return (
        <MethodContainer
            id={props.id}
            title="Password"
            explanation="Enter your password to confirm your identity"
            duoSelfEnrollment={false}
            registered={true}
            state={methodState}
        >
            <PasswordForm onAuthenticationSuccess={props.onAuthenticationSuccess} />
        </MethodContainer>
    );
};

export default PasswordMethod;
