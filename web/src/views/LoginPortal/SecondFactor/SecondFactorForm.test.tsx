import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import { SecondFactorMethod } from "@models/Methods";
import { AuthenticationLevel } from "@services/State";
import SecondFactorForm from "@views/LoginPortal/SecondFactor/SecondFactorForm";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@simplewebauthn/browser", () => ({
    browserSupportsWebAuthn: () => true,
}));

vi.mock("@contexts/LocalStorageMethodContext", () => ({
    useLocalStorageMethodContext: () => ({
        localStorageMethodAvailable: false,
        setLocalStorageMethod: vi.fn(),
    }),
}));

vi.mock("@hooks/Flow", () => ({
    useFlowPresent: () => false,
}));

vi.mock("@hooks/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
    }),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => vi.fn(),
}));

vi.mock("@constants/Routes", () => ({
    SecondFactorPasswordSubRoute: "/password",
    SecondFactorPushSubRoute: "/push",
    SecondFactorTOTPSubRoute: "/totp",
    SecondFactorWebAuthnSubRoute: "/webauthn",
    SettingsRoute: "/settings",
    SettingsTwoFactorAuthenticationSubRoute: "/2fa",
}));

vi.mock("@services/UserInfo", () => ({
    setPreferred2FAMethod: vi.fn(),
}));

vi.mock("@layouts/LoginLayout", () => ({
    default: (props: any) => <div data-testid="login-layout">{props.children}</div>,
}));

vi.mock("@components/LogoutButton", () => ({
    default: () => <button data-testid="logout-button">Logout</button>,
}));

vi.mock("@components/SwitchUserButton", () => ({
    default: () => <button data-testid="switch-user-button">Switch User</button>,
}));

vi.mock("@views/LoginPortal/SecondFactor/MethodSelectionDialog", () => ({
    default: () => <div data-testid="method-selection-dialog" />,
}));

const defaultProps = {
    authenticationLevel: AuthenticationLevel.OneFactor,
    configuration: { available_methods: new Set([SecondFactorMethod.TOTP, SecondFactorMethod.WebAuthn]) },
    duoSelfEnrollment: false,
    factorKnowledge: true,
    onAuthenticationSuccess: vi.fn(),
    onMethodChanged: vi.fn(),
    userInfo: {
        display_name: "John",
        has_duo: false,
        has_totp: true,
        has_webauthn: true,
    },
} as any;

beforeEach(() => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
});

it("renders login layout with user greeting", () => {
    render(
        <MemoryRouter>
            <SecondFactorForm {...defaultProps} />
        </MemoryRouter>,
    );
    expect(screen.getByTestId("login-layout")).toBeInTheDocument();
    expect(screen.getByTestId("logout-button")).toBeInTheDocument();
});

it("renders methods button when multiple methods available", () => {
    render(
        <MemoryRouter>
            <SecondFactorForm {...defaultProps} />
        </MemoryRouter>,
    );
    expect(screen.getByText("Methods")).toBeInTheDocument();
});

it("does not render methods button when only one method available", () => {
    const singleMethodProps = {
        ...defaultProps,
        configuration: { available_methods: new Set([SecondFactorMethod.TOTP]) },
    };
    render(
        <MemoryRouter>
            <SecondFactorForm {...singleMethodProps} />
        </MemoryRouter>,
    );
    expect(screen.queryByText("Methods")).not.toBeInTheDocument();
});
