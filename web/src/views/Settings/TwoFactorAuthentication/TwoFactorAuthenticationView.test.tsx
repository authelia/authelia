import { render, screen } from "@testing-library/react";

import { SecondFactorMethod } from "@models/Methods";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/LocalStorageMethodContext", () => ({
    useLocalStorageMethodContext: () => ({
        localStorageMethod: SecondFactorMethod.TOTP,
        localStorageMethodAvailable: false,
        setLocalStorageMethod: vi.fn(),
    }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: vi.fn(),
    }),
}));

const mockConfig = {
    available_methods: new Set([SecondFactorMethod.TOTP, SecondFactorMethod.WebAuthn]),
};

const mockUserInfo = {
    has_duo: false,
    has_telegram: false,
    has_totp: true,
    has_webauthn: true,
    method: SecondFactorMethod.TOTP,
};

vi.mock("@hooks/Configuration", () => ({
    useConfiguration: () => [mockConfig, vi.fn(), false, null],
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoPOST: () => [mockUserInfo, vi.fn(), false, null],
}));

vi.mock("@hooks/UserInfoTOTPConfiguration", () => ({
    useUserInfoTOTPConfigurationOptional: () => [{ digits: 6, period: 30 }, vi.fn(), false, null],
}));

vi.mock("@hooks/WebAuthnCredentials", () => ({
    useUserWebAuthnCredentials: () => [[], vi.fn(), false, null],
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordPanel", () => ({
    default: () => <div data-testid="otp-panel" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsPanel", () => ({
    default: () => <div data-testid="options-panel" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsDisabledPanel", () => ({
    default: () => <div data-testid="webauthn-disabled-panel" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsPanel", () => ({
    default: () => <div data-testid="webauthn-panel" />,
}));

it("renders OTP panel, WebAuthn panel, and options panel", () => {
    render(<TwoFactorAuthenticationView />);
    expect(screen.getByTestId("otp-panel")).toBeInTheDocument();
    expect(screen.getByTestId("webauthn-panel")).toBeInTheDocument();
    expect(screen.getByTestId("options-panel")).toBeInTheDocument();
});
