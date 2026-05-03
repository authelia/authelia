import { render, screen } from "@testing-library/react";

import PasskeyForm from "@views/LoginPortal/FirstFactor/PasskeyForm";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: null, id: null, subflow: null }),
}));

vi.mock("@models/WebAuthn", () => ({
    AssertionResult: { Success: 0 },
    AssertionResultFailureString: () => "failure",
}));

vi.mock("@services/WebAuthn", () => ({
    getWebAuthnPasskeyOptions: vi.fn(),
    getWebAuthnResult: vi.fn(),
    postWebAuthnPasskeyResponse: vi.fn(),
}));

vi.mock("@components/PasskeyIcon", () => ({
    default: () => <span data-testid="passkey-icon" />,
}));

it("renders passkey sign in button", () => {
    render(
        <PasskeyForm
            disabled={false}
            rememberMe={false}
            onAuthenticationStart={vi.fn()}
            onAuthenticationStop={vi.fn()}
            onAuthenticationError={vi.fn()}
            onAuthenticationSuccess={vi.fn()}
        />,
    );
    expect(screen.getByText("Sign in with a passkey")).toBeInTheDocument();
    expect(screen.getByText("or")).toBeInTheDocument();
});

it("renders button as disabled when disabled prop is true", () => {
    render(
        <PasskeyForm
            disabled={true}
            rememberMe={false}
            onAuthenticationStart={vi.fn()}
            onAuthenticationStop={vi.fn()}
            onAuthenticationError={vi.fn()}
            onAuthenticationSuccess={vi.fn()}
        />,
    );
    expect(screen.getByText("Sign in with a passkey").closest("button")).toBeDisabled();
});
