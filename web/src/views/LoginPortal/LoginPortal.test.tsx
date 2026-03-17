import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import { useLocalStorageMethodContext } from "@contexts/LocalStorageMethodContext";
import { useNotifications } from "@contexts/NotificationsContext";
import { useConfiguration } from "@hooks/Configuration";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import { useUserInfoPOST } from "@hooks/UserInfo";
import LoginPortal from "@views/LoginPortal/LoginPortal";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@constants/Routes", () => ({
    AuthenticatedRoute: "/authenticated",
    IndexRoute: "/",
    SecondFactorPasswordSubRoute: "/password",
    SecondFactorPushSubRoute: "/push",
    SecondFactorRoute: "/2fa",
    SecondFactorTOTPSubRoute: "/totp",
    SecondFactorWebAuthnSubRoute: "/webauthn",
}));

vi.mock("@contexts/LocalStorageMethodContext", () => ({
    useLocalStorageMethodContext: vi.fn(),
}));

vi.mock("@hooks/Configuration", () => ({
    useConfiguration: vi.fn(),
}));

vi.mock("@contexts/NotificationsContext", () => ({
    useNotifications: vi.fn(),
}));

vi.mock("@hooks/QueryParam", () => ({
    useQueryParam: () => null,
}));

vi.mock("@hooks/Redirector", () => ({
    useRedirector: () => vi.fn(),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: vi.fn(),
}));

vi.mock("@hooks/State", () => ({
    useAutheliaState: vi.fn(),
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoPOST: vi.fn(),
}));

vi.mock("@services/SafeRedirection", () => ({
    checkSafeRedirection: vi.fn(),
}));

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div data-testid="loading-page" />,
}));

vi.mock("@views/LoginPortal/AuthenticatedView/AuthenticatedView", () => ({
    default: () => <div data-testid="authenticated-view" />,
}));

vi.mock("@views/LoginPortal/FirstFactor/FirstFactorForm", () => ({
    default: () => <div data-testid="first-factor-form" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/SecondFactorForm", () => ({
    default: () => <div data-testid="second-factor-form" />,
}));

const mockNavigate = vi.fn();
const mockCreateErrorNotification = vi.fn();

const defaultProps = {
    duoSelfEnrollment: false,
    passkeyLogin: false,
    rememberMe: true,
    resetPassword: true,
    resetPasswordCustomURL: "",
};

const mockNotificationsReturn: ReturnType<typeof useNotifications> = {
    createErrorNotification: mockCreateErrorNotification,
    createInfoNotification: vi.fn(),
    createSuccessNotification: vi.fn(),
    createWarnNotification: vi.fn(),
    isActive: false,
    notification: null,
    resetNotification: vi.fn(),
};

beforeEach(() => {
    vi.mocked(useRouterNavigate).mockReturnValue(mockNavigate);
    vi.mocked(useNotifications).mockReturnValue(mockNotificationsReturn);
    vi.mocked(useLocalStorageMethodContext).mockReturnValue({
        localStorageMethod: undefined,
        localStorageMethodAvailable: false,
        setLocalStorageMethod: vi.fn(),
    });
    vi.mocked(useAutheliaState).mockReturnValue([undefined, vi.fn(), false, undefined]);
    vi.mocked(useConfiguration).mockReturnValue([undefined, vi.fn(), false, undefined]);
    vi.mocked(useUserInfoPOST).mockReturnValue([undefined, vi.fn(), false, undefined]);
    mockNavigate.mockClear();
    mockCreateErrorNotification.mockClear();
});

it("renders loading page when state is not loaded", () => {
    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );
    expect(screen.getByTestId("loading-page")).toBeInTheDocument();
});

it("unauthenticated state navigates to IndexRoute", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 0, factor_knowledge: false, username: "" },
        vi.fn(),
        false,
        undefined,
    ]);

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalled();
    });
    expect(mockNavigate).toHaveBeenCalledWith("/");
});

it("OneFactor with no 2FA methods navigates to /authenticated", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useConfiguration).mockReturnValue([
        { available_methods: new Set(), password_change_disabled: false, password_reset_disabled: false },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserInfoPOST).mockReturnValue([
        { display_name: "test", emails: [], has_duo: false, has_totp: false, has_webauthn: false, method: 1 },
        vi.fn(),
        false,
        undefined,
    ]);

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledTimes(1);
    });
    expect(mockNavigate).toHaveBeenNthCalledWith(1, "/authenticated", false);
});

it("OneFactor with TOTP preferred navigates to /2fa/totp", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useConfiguration).mockReturnValue([
        { available_methods: new Set([1]), password_change_disabled: false, password_reset_disabled: false },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserInfoPOST).mockReturnValue([
        { display_name: "test", emails: [], has_duo: false, has_totp: true, has_webauthn: false, method: 1 },
        vi.fn(),
        false,
        undefined,
    ]);

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledTimes(1);
    });
    expect(mockNavigate).toHaveBeenNthCalledWith(1, "/2fa/totp");
});

it("OneFactor with WebAuthn preferred navigates to /2fa/webauthn", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useConfiguration).mockReturnValue([
        { available_methods: new Set([2]), password_change_disabled: false, password_reset_disabled: false },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserInfoPOST).mockReturnValue([
        { display_name: "test", emails: [], has_duo: false, has_totp: false, has_webauthn: true, method: 2 },
        vi.fn(),
        false,
        undefined,
    ]);

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledTimes(1);
    });
    expect(mockNavigate).toHaveBeenNthCalledWith(1, "/2fa/webauthn");
});

it("OneFactor with MobilePush preferred navigates to /2fa/push", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useConfiguration).mockReturnValue([
        { available_methods: new Set([3]), password_change_disabled: false, password_reset_disabled: false },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserInfoPOST).mockReturnValue([
        { display_name: "test", emails: [], has_duo: true, has_totp: false, has_webauthn: false, method: 3 },
        vi.fn(),
        false,
        undefined,
    ]);

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledTimes(1);
    });
    expect(mockNavigate).toHaveBeenNthCalledWith(1, "/2fa/push");
});

it("OneFactor with factor_knowledge false navigates to /2fa/password", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: false, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useConfiguration).mockReturnValue([
        { available_methods: new Set([1]), password_change_disabled: false, password_reset_disabled: false },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserInfoPOST).mockReturnValue([
        { display_name: "test", emails: [], has_duo: false, has_totp: true, has_webauthn: false, method: 1 },
        vi.fn(),
        false,
        undefined,
    ]);

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledTimes(1);
    });
    expect(mockNavigate).toHaveBeenNthCalledWith(1, "/2fa/password");
});

it("localStorageMethod overrides userInfo.method", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useConfiguration).mockReturnValue([
        { available_methods: new Set([1, 2]), password_change_disabled: false, password_reset_disabled: false },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserInfoPOST).mockReturnValue([
        { display_name: "test", emails: [], has_duo: false, has_totp: true, has_webauthn: false, method: 1 },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useLocalStorageMethodContext).mockReturnValue({
        localStorageMethod: 2,
        localStorageMethodAvailable: true,
        setLocalStorageMethod: vi.fn(),
    });

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledTimes(1);
    });
    expect(mockNavigate).toHaveBeenNthCalledWith(1, "/2fa/webauthn");
});

it("fetchStateError triggers createErrorNotification", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([undefined, vi.fn(), false, new Error("state error")]);

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockCreateErrorNotification).toHaveBeenCalledTimes(1);
    });
});

it("fetchConfigurationError triggers createErrorNotification", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useConfiguration).mockReturnValue([undefined, vi.fn(), false, new Error("config error")]);

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockCreateErrorNotification).toHaveBeenCalledTimes(1);
    });
});

it("fetchUserInfoError triggers createErrorNotification", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserInfoPOST).mockReturnValue([undefined, vi.fn(), false, new Error("userinfo error")]);

    render(
        <MemoryRouter>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => {
        expect(mockCreateErrorNotification).toHaveBeenCalledTimes(1);
    });
});
