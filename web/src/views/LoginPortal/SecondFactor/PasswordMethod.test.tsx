import { render, screen } from "@testing-library/react";

import { AuthenticationLevel } from "@services/State";
import PasswordMethod from "@views/LoginPortal/SecondFactor/PasswordMethod";

vi.mock("@views/LoginPortal/SecondFactor/MethodContainer", () => ({
    default: (props: any) => (
        <div data-testid="method-container" data-state={props.state} data-title={props.title}>
            {props.children}
        </div>
    ),
    State: { ALREADY_AUTHENTICATED: 0, METHOD: 2, NOT_REGISTERED: 1 },
}));

vi.mock("@views/LoginPortal/SecondFactor/PasswordForm", () => ({
    default: () => <div data-testid="password-form" />,
}));

it("renders in METHOD state when not two-factor authenticated", () => {
    render(
        <PasswordMethod
            id="test"
            authenticationLevel={AuthenticationLevel.OneFactor}
            onAuthenticationSuccess={vi.fn()}
        />,
    );
    expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "2");
    expect(screen.getByTestId("password-form")).toBeInTheDocument();
});

it("renders in ALREADY_AUTHENTICATED state when two-factor authenticated", () => {
    render(
        <PasswordMethod
            id="test"
            authenticationLevel={AuthenticationLevel.TwoFactor}
            onAuthenticationSuccess={vi.fn()}
        />,
    );
    expect(screen.getByTestId("method-container")).toHaveAttribute("data-state", "0");
});
