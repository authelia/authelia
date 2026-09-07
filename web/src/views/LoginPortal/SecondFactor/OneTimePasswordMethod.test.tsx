import { render, screen } from "@testing-library/react";

import { AuthenticationLevel } from "@services/State";
import OneTimePasswordMethod from "@views/LoginPortal/SecondFactor/OneTimePasswordMethod";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: null, id: null, subflow: null }),
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => null,
}));

vi.mock("@hooks/UserInfoTOTPConfiguration", () => ({
    useUserInfoTOTPConfiguration: () => [{ digits: 6, period: 30 }, vi.fn(), false, null],
}));

vi.mock("@services/OneTimePassword", () => ({
    completeTOTPSignIn: vi.fn(),
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/OTPDial", () => ({
    default: (props: any) => <div data-testid="otp-dial" data-digits={props.digits} data-period={props.period} />,
    State: { Failure: 3, Idle: 0, InProgress: 1, RateLimited: 4, Success: 2 },
}));

vi.mock("@views/LoginPortal/SecondFactor/MethodContainer", () => ({
    default: (props: any) => (
        <div data-testid="method-container" data-state={props.state} data-title={props.title}>
            {props.children}
        </div>
    ),
    State: { ALREADY_AUTHENTICATED: "ALREADY_AUTHENTICATED", METHOD: "METHOD", NOT_REGISTERED: "NOT_REGISTERED" },
}));

it("renders method container with OTP dial when registered", () => {
    render(
        <OneTimePasswordMethod
            id="otp-method"
            authenticationLevel={AuthenticationLevel.OneFactor}
            registered={true}
            onRegisterClick={vi.fn()}
            onSignInError={vi.fn()}
            onSignInSuccess={vi.fn()}
        />,
    );
    expect(screen.getByTestId("method-container")).toHaveAttribute("data-title", "One-Time Password");
    expect(screen.getByTestId("otp-dial")).toBeInTheDocument();
});

it("renders not registered state when not registered", () => {
    render(
        <OneTimePasswordMethod
            id="otp-method"
            authenticationLevel={AuthenticationLevel.OneFactor}
            registered={false}
            onRegisterClick={vi.fn()}
            onSignInError={vi.fn()}
            onSignInSuccess={vi.fn()}
        />,
    );
    expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "NOT_REGISTERED");
});

it("renders already authenticated state at two factor level", () => {
    render(
        <OneTimePasswordMethod
            id="otp-method"
            authenticationLevel={AuthenticationLevel.TwoFactor}
            registered={true}
            onRegisterClick={vi.fn()}
            onSignInError={vi.fn()}
            onSignInSuccess={vi.fn()}
        />,
    );
    expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "ALREADY_AUTHENTICATED");
});
