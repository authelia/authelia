// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { SecondFactorMethod } from "@models/Methods";
import TwoFactorAuthenticationView from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView";

const mocks = vi.hoisted(() => ({
    configuration: undefined as any,
    configurationError: null as Error | null,
    createErrorNotification: vi.fn(),
    credentials: [] as any,
    credentialsError: null as Error | null,
    fetchConfiguration: vi.fn(),
    fetchCredentials: vi.fn(),
    fetchTOTPConfig: vi.fn(),
    fetchUserInfo: vi.fn(),
    localStorageMethodAvailable: false,
    setLocalStorageMethod: vi.fn(),
    totpConfig: { digits: 6, period: 30 } as any,
    totpConfigError: null as Error | null,
    userInfo: undefined as any,
    userInfoError: null as Error | null,
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@contexts/LocalStorageMethodContext", () => ({
    useLocalStorageMethodContext: () => ({
        localStorageMethod: SecondFactorMethod.TOTP,
        localStorageMethodAvailable: mocks.localStorageMethodAvailable,
        setLocalStorageMethod: mocks.setLocalStorageMethod,
    }),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: () => ({
        createErrorNotification: mocks.createErrorNotification,
    }),
}));

vi.mock("@hooks/Configuration", () => ({
    useConfiguration: () => [mocks.configuration, mocks.fetchConfiguration, false, mocks.configurationError],
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoPOST: () => [mocks.userInfo, mocks.fetchUserInfo, false, mocks.userInfoError],
}));

vi.mock("@hooks/UserInfoTOTPConfiguration", () => ({
    useUserInfoTOTPConfigurationOptional: () => [mocks.totpConfig, mocks.fetchTOTPConfig, false, mocks.totpConfigError],
}));

vi.mock("@hooks/WebAuthnCredentials", () => ({
    useUserWebAuthnCredentials: () => [mocks.credentials, mocks.fetchCredentials, false, mocks.credentialsError],
}));

vi.mock("@views/Settings/TwoFactorAuthentication/OneTimePasswordPanel", () => ({
    default: (props: any) => (
        <div data-testid="otp-panel">
            <button data-testid="otp-refresh" onClick={() => props.handleRefreshState()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsPanel", () => ({
    default: (props: any) => (
        <div data-testid="options-panel">
            <button data-testid="options-refresh" onClick={() => props.refresh()} />
        </div>
    ),
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsDisabledPanel", () => ({
    default: () => <div data-testid="webauthn-disabled-panel" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/WebAuthnCredentialsPanel", () => ({
    default: (props: any) => (
        <div data-testid="webauthn-panel">
            <button data-testid="webauthn-refresh" onClick={() => props.handleRefreshState()} />
        </div>
    ),
}));

const bothMethods = { available_methods: new Set([SecondFactorMethod.TOTP, SecondFactorMethod.WebAuthn]) };
const userInfo = {
    has_duo: false,
    has_totp: true,
    has_webauthn: true,
    method: SecondFactorMethod.TOTP,
};

beforeEach(() => {
    vi.clearAllMocks();
    mocks.configuration = bothMethods;
    mocks.configurationError = null;
    mocks.credentials = [];
    mocks.credentialsError = null;
    mocks.localStorageMethodAvailable = false;
    mocks.totpConfig = { digits: 6, period: 30 };
    mocks.totpConfigError = null;
    mocks.userInfo = userInfo;
    mocks.userInfoError = null;
});

describe("rendering", () => {
    it("renders OTP panel, WebAuthn panel, and options panel", () => {
        render(<TwoFactorAuthenticationView />);
        expect(screen.getByTestId("otp-panel")).toBeInTheDocument();
        expect(screen.getByTestId("webauthn-panel")).toBeInTheDocument();
        expect(screen.getByTestId("options-panel")).toBeInTheDocument();
    });

    it("fetches the configuration and user info on mount", () => {
        render(<TwoFactorAuthenticationView />);
        expect(mocks.fetchConfiguration).toHaveBeenCalled();
        expect(mocks.fetchUserInfo).toHaveBeenCalled();
    });

    it("hides the OTP panel when TOTP is not an available method", () => {
        mocks.configuration = { available_methods: new Set([SecondFactorMethod.WebAuthn]) };

        render(<TwoFactorAuthenticationView />);

        expect(screen.queryByTestId("otp-panel")).not.toBeInTheDocument();
        expect(screen.getByTestId("webauthn-panel")).toBeInTheDocument();
    });

    it("shows the disabled WebAuthn panel when WebAuthn is not an available method", () => {
        mocks.configuration = { available_methods: new Set([SecondFactorMethod.TOTP]) };

        render(<TwoFactorAuthenticationView />);

        expect(screen.getByTestId("webauthn-disabled-panel")).toBeInTheDocument();
        expect(screen.queryByTestId("webauthn-panel")).not.toBeInTheDocument();
    });

    it("explains that no second factor is required when no methods are available", () => {
        mocks.configuration = { available_methods: new Set() };

        render(<TwoFactorAuthenticationView />);

        expect(
            screen.getByText("There are no protected applications that require a second factor method"),
        ).toBeInTheDocument();
        expect(screen.queryByTestId("otp-panel")).not.toBeInTheDocument();
        expect(screen.queryByTestId("webauthn-panel")).not.toBeInTheDocument();
        expect(screen.queryByTestId("webauthn-disabled-panel")).not.toBeInTheDocument();
    });

    it("hides the options panel while the configuration is unavailable", () => {
        mocks.configuration = undefined;

        render(<TwoFactorAuthenticationView />);

        expect(screen.queryByTestId("options-panel")).not.toBeInTheDocument();
    });

    it("hides the options panel while the user info is unavailable", () => {
        mocks.userInfo = undefined;

        render(<TwoFactorAuthenticationView />);

        expect(screen.queryByTestId("options-panel")).not.toBeInTheDocument();
    });
});

describe("credential fetching", () => {
    it("fetches the TOTP configuration when TOTP is enabled", () => {
        render(<TwoFactorAuthenticationView />);
        expect(mocks.fetchTOTPConfig).toHaveBeenCalled();
    });

    it("does not fetch the TOTP configuration when TOTP is disabled", () => {
        mocks.configuration = { available_methods: new Set([SecondFactorMethod.WebAuthn]) };

        render(<TwoFactorAuthenticationView />);

        expect(mocks.fetchTOTPConfig).not.toHaveBeenCalled();
    });

    it("fetches the WebAuthn credentials when WebAuthn is enabled", () => {
        render(<TwoFactorAuthenticationView />);
        expect(mocks.fetchCredentials).toHaveBeenCalled();
    });

    it("does not fetch the WebAuthn credentials when WebAuthn is disabled", () => {
        mocks.configuration = { available_methods: new Set([SecondFactorMethod.TOTP]) };

        render(<TwoFactorAuthenticationView />);

        expect(mocks.fetchCredentials).not.toHaveBeenCalled();
    });
});

describe("local storage method", () => {
    it("selects the sole available method when local storage is available", () => {
        mocks.localStorageMethodAvailable = true;
        mocks.configuration = { available_methods: new Set([SecondFactorMethod.TOTP]) };

        render(<TwoFactorAuthenticationView />);

        expect(mocks.setLocalStorageMethod).toHaveBeenCalledWith(SecondFactorMethod.TOTP);
    });

    it("does not select a method when several are available", () => {
        mocks.localStorageMethodAvailable = true;

        render(<TwoFactorAuthenticationView />);

        expect(mocks.setLocalStorageMethod).not.toHaveBeenCalled();
    });

    it("does not select a method when local storage is unavailable", () => {
        mocks.localStorageMethodAvailable = false;
        mocks.configuration = { available_methods: new Set([SecondFactorMethod.TOTP]) };

        render(<TwoFactorAuthenticationView />);

        expect(mocks.setLocalStorageMethod).not.toHaveBeenCalled();
    });
});

describe("fetch errors", () => {
    it("notifies when the configuration cannot be retrieved", () => {
        mocks.configurationError = new Error("boom");

        render(<TwoFactorAuthenticationView />);

        expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue retrieving the {{item}}");
    });

    it("notifies when the user info cannot be retrieved", () => {
        mocks.userInfoError = new Error("boom");

        render(<TwoFactorAuthenticationView />);

        expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue retrieving the {{item}}");
    });

    it("notifies when the TOTP configuration cannot be retrieved", () => {
        mocks.totpConfigError = new Error("boom");

        render(<TwoFactorAuthenticationView />);

        expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue retrieving the {{item}}");
    });

    it("notifies when the WebAuthn credentials cannot be retrieved", () => {
        mocks.credentialsError = new Error("boom");

        render(<TwoFactorAuthenticationView />);

        expect(mocks.createErrorNotification).toHaveBeenCalledWith("There was an issue retrieving the {{item}}");
    });
});

describe("refreshing", () => {
    it("refetches everything when the TOTP panel asks for a refresh", async () => {
        render(<TwoFactorAuthenticationView />);

        mocks.fetchConfiguration.mockClear();
        mocks.fetchUserInfo.mockClear();
        mocks.fetchTOTPConfig.mockClear();

        fireEvent.click(screen.getByTestId("otp-refresh"));

        await waitFor(() => expect(mocks.fetchTOTPConfig).toHaveBeenCalled());
        expect(mocks.fetchConfiguration).toHaveBeenCalled();
        expect(mocks.fetchUserInfo).toHaveBeenCalled();
    });

    it("refetches everything when the WebAuthn panel asks for a refresh", async () => {
        render(<TwoFactorAuthenticationView />);

        mocks.fetchConfiguration.mockClear();
        mocks.fetchUserInfo.mockClear();
        mocks.fetchCredentials.mockClear();

        fireEvent.click(screen.getByTestId("webauthn-refresh"));

        await waitFor(() => expect(mocks.fetchCredentials).toHaveBeenCalled());
        expect(mocks.fetchConfiguration).toHaveBeenCalled();
        expect(mocks.fetchUserInfo).toHaveBeenCalled();
    });

    it("refetches the user info when the options panel asks for a refresh", async () => {
        render(<TwoFactorAuthenticationView />);

        mocks.fetchUserInfo.mockClear();

        fireEvent.click(screen.getByTestId("options-refresh"));

        await waitFor(() => expect(mocks.fetchUserInfo).toHaveBeenCalled());
    });
});
