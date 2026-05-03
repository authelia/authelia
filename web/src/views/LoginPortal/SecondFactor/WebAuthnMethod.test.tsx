import { render, screen } from "@testing-library/react";

import { AuthenticationLevel } from "@services/State";
import WebAuthnMethod from "@views/LoginPortal/SecondFactor/WebAuthnMethod";

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

vi.mock("@models/WebAuthn", () => ({
    AssertionResult: { Success: 0 },
    AssertionResultFailureString: () => "failure",
    WebAuthnTouchState: { Failure: 2, InProgress: 1, WaitTouch: 0 },
}));

vi.mock("@services/WebAuthn", () => ({
    getWebAuthnOptions: vi.fn().mockResolvedValue({ options: null, status: 400 }),
    getWebAuthnResult: vi.fn(),
    postWebAuthnResponse: vi.fn(),
}));

vi.mock("@components/WebAuthnTryIcon", () => ({
    default: (props: any) => <div data-testid="webauthn-try-icon" data-state={props.webauthnTouchState} />,
}));

vi.mock("@views/LoginPortal/SecondFactor/MethodContainer", () => ({
    default: (props: any) => (
        <div data-testid="method-container" data-state={props.state} data-title={props.title}>
            {props.children}
        </div>
    ),
    State: { ALREADY_AUTHENTICATED: "ALREADY_AUTHENTICATED", METHOD: "METHOD", NOT_REGISTERED: "NOT_REGISTERED" },
}));

it("renders method container with security key title", () => {
    vi.spyOn(console, "error").mockImplementation(() => {});
    render(
        <WebAuthnMethod
            id="webauthn-method"
            authenticationLevel={AuthenticationLevel.OneFactor}
            registered={true}
            onRegisterClick={vi.fn()}
            onSignInError={vi.fn()}
            onSignInSuccess={vi.fn()}
        />,
    );
    expect(screen.getByTestId("method-container")).toHaveAttribute("data-title", "Security Key");
    expect(screen.getByTestId("webauthn-try-icon")).toBeInTheDocument();
});

it("renders not registered state when not registered", () => {
    render(
        <WebAuthnMethod
            id="webauthn-method"
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
        <WebAuthnMethod
            id="webauthn-method"
            authenticationLevel={AuthenticationLevel.TwoFactor}
            registered={true}
            onRegisterClick={vi.fn()}
            onSignInError={vi.fn()}
            onSignInSuccess={vi.fn()}
        />,
    );
    expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "ALREADY_AUTHENTICATED");
});
