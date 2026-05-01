import { render, screen } from "@testing-library/react";

import FirstFactorForm from "@views/LoginPortal/FirstFactor/FirstFactorForm";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("react-router-dom", async () => {
    const actual = await vi.importActual("react-router-dom");
    return { ...actual, useNavigate: () => vi.fn() };
});

vi.mock("broadcast-channel", () => {
    class MockBroadcastChannel {
        addEventListener = vi.fn();
        removeEventListener = vi.fn();
        postMessage = vi.fn();
    }
    return { BroadcastChannel: MockBroadcastChannel };
});

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Flow", () => ({
    useFlow: () => ({ flow: null, id: null, subflow: null }),
}));

vi.mock("@hooks/OpenIDConnect", () => ({
    useUserCode: () => null,
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
    }),
}));

vi.mock("@layouts/LoginLayout", () => ({
    default: (props: any) => <div data-testid="login-layout">{props.children}</div>,
}));

vi.mock("@services/CapsLock", () => ({
    IsCapsLockModified: () => null,
}));

vi.mock("@services/Password", () => ({
    postFirstFactor: vi.fn(),
}));

vi.mock("@views/LoginPortal/FirstFactor/PasskeyForm", () => ({
    default: () => <div data-testid="passkey-form" />,
}));

const defaultProps = {
    disabled: false,
    onAuthenticationStart: vi.fn(),
    onAuthenticationStop: vi.fn(),
    onAuthenticationSuccess: vi.fn(),
    onChannelStateChange: vi.fn(),
    passkeyLogin: false,
    rememberMe: false,
    resetPassword: false,
    resetPasswordCustomURL: "",
};

it("renders login form with username, password, and sign in button", () => {
    render(<FirstFactorForm {...defaultProps} />);
    expect(screen.getByText("Username")).toBeInTheDocument();
    expect(screen.getByText("Password")).toBeInTheDocument();
    expect(screen.getByText("Sign in")).toBeInTheDocument();
});

it("renders remember me checkbox when enabled", () => {
    render(<FirstFactorForm {...defaultProps} rememberMe={true} />);
    expect(screen.getByText("Remember me")).toBeInTheDocument();
});

it("does not render remember me checkbox when disabled", () => {
    render(<FirstFactorForm {...defaultProps} rememberMe={false} />);
    expect(screen.queryByText("Remember me")).not.toBeInTheDocument();
});

it("renders passkey form when passkey login is enabled", () => {
    render(<FirstFactorForm {...defaultProps} passkeyLogin={true} />);
    expect(screen.getByTestId("passkey-form")).toBeInTheDocument();
});

it("renders reset password link when enabled", () => {
    render(<FirstFactorForm {...defaultProps} resetPassword={true} />);
    expect(screen.getByText("Reset password?")).toBeInTheDocument();
});

it("does not render reset password link when disabled", () => {
    render(<FirstFactorForm {...defaultProps} resetPassword={false} />);
    expect(screen.queryByText("Reset password?")).not.toBeInTheDocument();
});
