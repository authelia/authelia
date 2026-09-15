// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import { useLocalStorageMethodContext } from "@contexts/LocalStorageMethodContext";
import { useNotifications } from "@contexts/NotificationsContext";
import { useConfiguration } from "@hooks/Configuration";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useAutheliaState } from "@hooks/State";
import { useUserInfoPOST } from "@hooks/UserInfo";
import { checkSafeRedirection } from "@services/SafeRedirection";
import LoginPortal from "@views/LoginPortal/LoginPortal";

const mocks = vi.hoisted(() => ({
    fetchState: vi.fn(),
    fetchUserInfo: vi.fn(),
    redirectionURL: null as null | string,
    redirector: vi.fn(),
}));

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@constants/Routes", () => ({
    AuthenticatedRoute: "/authenticated",
    ExternalIdentityLinkRoute: "/external-identity/link",
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
    useQueryParam: (param: string) => (param === "rd" ? (mocks.redirectionURL ?? undefined) : mockQueryParams[param]),
}));

vi.mock("@hooks/Redirector", () => ({
    useRedirector: () => mocks.redirector,
}));

vi.mock("@services/ExternalIdentity", () => ({
    postExternalIdentityStart: (...args: unknown[]) => mockPostOpenIDConnectStart(...args),
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
    default: (props: any) => (
        <div
            data-testid="first-factor-form"
            data-disabled={String(props.disabled)}
            data-openid-connect-login={String(props.externalIdentityLogin)}
            data-external-identity-link={props.externalIdentityLink ?? ""}
        >
            <button data-testid="ff-start" onClick={() => props.onAuthenticationStart()} />
            <button data-testid="ff-stop" onClick={() => props.onAuthenticationStop()} />
            <button
                data-testid="ff-success-redirect"
                onClick={() => props.onAuthenticationSuccess("https://example.com")}
            />
            <button data-testid="ff-success-plain" onClick={() => props.onAuthenticationSuccess(undefined)} />
            <button data-testid="ff-channel" onClick={() => props.onChannelStateChange()} />
        </div>
    ),
}));

vi.mock("@views/LoginPortal/SecondFactor/SecondFactorForm", () => ({
    default: (props: any) => (
        <div data-testid="second-factor-form" data-level={props.authenticationLevel}>
            <button data-testid="sf-method-changed" onClick={() => props.onMethodChanged()} />
            <button
                data-testid="sf-success-redirect"
                onClick={() => props.onAuthenticationSuccess("https://example.com")}
            />
            <button data-testid="sf-success-plain" onClick={() => props.onAuthenticationSuccess(undefined)} />
        </div>
    ),
}));

const mockNavigate = vi.fn();
const mockCreateErrorNotification = vi.fn();
const mockCreateInfoNotification = vi.fn();
const mockPostOpenIDConnectStart = vi.fn();

let mockQueryParams: Record<string, string | undefined> = {};

const defaultProps = {
    duoSelfEnrollment: false,
    externalIdentityLogin: false,
    passkeyLogin: false,
    rememberMe: true,
    resetPassword: true,
    resetPasswordCustomURL: "",
};

const mockNotificationsReturn: ReturnType<typeof useNotifications> = {
    createErrorNotification: mockCreateErrorNotification,
    createInfoNotification: mockCreateInfoNotification,
    createSuccessNotification: vi.fn(),
    createWarnNotification: vi.fn(),
    isActive: false,
    notification: null,
    resetNotification: vi.fn(),
    showNotification: vi.fn(),
};

beforeEach(() => {
    vi.mocked(useRouterNavigate).mockReturnValue(mockNavigate);
    vi.mocked(useNotifications).mockReturnValue(mockNotificationsReturn);
    vi.mocked(useLocalStorageMethodContext).mockReturnValue({
        localStorageMethod: undefined,
        localStorageMethodAvailable: false,
        setLocalStorageMethod: vi.fn(),
    });
    vi.mocked(useAutheliaState).mockReturnValue([undefined, mocks.fetchState, false, undefined]);
    vi.mocked(useConfiguration).mockReturnValue([undefined, vi.fn(), false, undefined]);
    vi.mocked(useUserInfoPOST).mockReturnValue([undefined, mocks.fetchUserInfo, false, undefined]);
    mocks.redirectionURL = null;
    mocks.redirector.mockClear();
    mocks.fetchState.mockClear();
    mocks.fetchUserInfo.mockClear();
    vi.mocked(checkSafeRedirection).mockReset();
    mockNavigate.mockClear();
    mockCreateErrorNotification.mockClear();
    mockCreateInfoNotification.mockClear();
    mockPostOpenIDConnectStart.mockReset();
    mockQueryParams = {};
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

it("resumes the external identity link on the link page", async () => {
    mockQueryParams = { link_provider: "google" };
    mockPostOpenIDConnectStart.mockResolvedValue({ authorization_url: "https://op.example.com/authorize?x=1" });

    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 2, factor_knowledge: true, username: "test" },
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
        <MemoryRouter initialEntries={["/external-identity/link"]}>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => expect(mocks.redirector).toHaveBeenCalledWith("https://op.example.com/authorize?x=1"));

    expect(mockPostOpenIDConnectStart).toHaveBeenCalledWith("google", {
        keepMeLoggedIn: false,
        requestMethod: undefined,
        targetURL: undefined,
    });
    expect(mockNavigate).not.toHaveBeenCalled();
});

it("does not resume the OpenID Connect linking flow when the parameter is absent", async () => {
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 2, factor_knowledge: true, username: "test" },
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

    await waitFor(() => expect(mockNavigate).toHaveBeenCalled());

    expect(mockPostOpenIDConnectStart).not.toHaveBeenCalled();
    expect(mocks.redirector).not.toHaveBeenCalled();
});

it("does not resume the external identity link while a second factor is owed", async () => {
    mockQueryParams = { link_provider: "google" };
    mockPostOpenIDConnectStart.mockResolvedValue({ authorization_url: "https://op.example.com/authorize?x=1" });

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
        <MemoryRouter initialEntries={["/external-identity/link"]}>
            <LoginPortal {...defaultProps} />
        </MemoryRouter>,
    );

    await waitFor(() => expect(mockNavigate).toHaveBeenNthCalledWith(1, "/2fa/totp"));

    expect(mockPostOpenIDConnectStart).not.toHaveBeenCalled();
    expect(mocks.redirector).not.toHaveBeenCalled();
});

it("ignores the external identity link parameter on the index page", async () => {
    mockQueryParams = { link_provider: "google" };

    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 2, factor_knowledge: true, username: "test" },
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

    await waitFor(() => expect(mockNavigate).toHaveBeenCalled());

    expect(mockPostOpenIDConnectStart).not.toHaveBeenCalled();
    expect(mocks.redirector).not.toHaveBeenCalled();
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

describe("safe redirection", () => {
    const unauthenticated = { authentication_level: 0, factor_knowledge: false, username: "test" };
    const twoFactor = { authentication_level: 2, factor_knowledge: true, username: "test" };
    const oneFactor = { authentication_level: 1, factor_knowledge: true, username: "test" };

    const userInfo = {
        display_name: "test",
        emails: [],
        has_duo: false,
        has_totp: true,
        has_webauthn: false,
        method: 1,
    };

    function renderPortal(route = "/") {
        return render(
            <MemoryRouter initialEntries={[route]}>
                <LoginPortal {...defaultProps} />
            </MemoryRouter>,
        );
    }

    it("redirects to a safe target once fully authenticated", async () => {
        mocks.redirectionURL = "https://app.example.com";
        vi.mocked(checkSafeRedirection).mockResolvedValue({ ok: true } as any);
        vi.mocked(useAutheliaState).mockReturnValue([twoFactor, mocks.fetchState, false, undefined]);

        renderPortal();

        await waitFor(() => expect(mocks.redirector).toHaveBeenCalledWith("https://app.example.com"));
        expect(mockNavigate).not.toHaveBeenCalled();
    });

    it("notifies when the target is rejected as unsafe", async () => {
        mocks.redirectionURL = "https://evil.example.com";
        vi.mocked(checkSafeRedirection).mockResolvedValue({ ok: false } as any);
        vi.mocked(useAutheliaState).mockReturnValue([twoFactor, mocks.fetchState, false, undefined]);

        renderPortal();

        await waitFor(() => expect(mockCreateErrorNotification).toHaveBeenCalled());
        expect(mocks.redirector).not.toHaveBeenCalled();
    });

    it("notifies when the safety check throws", async () => {
        mocks.redirectionURL = "https://app.example.com";
        vi.mocked(checkSafeRedirection).mockRejectedValue(new Error("boom"));
        vi.mocked(useAutheliaState).mockReturnValue([twoFactor, mocks.fetchState, false, undefined]);

        renderPortal();

        await waitFor(() => expect(mockCreateErrorNotification).toHaveBeenCalled());
        expect(mocks.redirector).not.toHaveBeenCalled();
    });

    it("redirects at one factor level when no second factor methods exist", async () => {
        mocks.redirectionURL = "https://app.example.com";
        vi.mocked(checkSafeRedirection).mockResolvedValue({ ok: true } as any);
        vi.mocked(useAutheliaState).mockReturnValue([oneFactor, mocks.fetchState, false, undefined]);
        vi.mocked(useConfiguration).mockReturnValue([
            { available_methods: new Set(), password_change_disabled: false, password_reset_disabled: false },
            vi.fn(),
            false,
            undefined,
        ]);
        vi.mocked(useUserInfoPOST).mockReturnValue([userInfo, mocks.fetchUserInfo, false, undefined]);

        renderPortal();

        await waitFor(() => expect(mocks.redirector).toHaveBeenCalledWith("https://app.example.com"));
    });

    it("does not redirect while still unauthenticated", async () => {
        mocks.redirectionURL = "https://app.example.com";
        vi.mocked(checkSafeRedirection).mockResolvedValue({ ok: true } as any);
        vi.mocked(useAutheliaState).mockReturnValue([unauthenticated, mocks.fetchState, false, undefined]);

        renderPortal();

        await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith("/"));
        expect(mocks.redirector).not.toHaveBeenCalled();
    });

    it("does not check safety without a redirection URL", async () => {
        vi.mocked(useAutheliaState).mockReturnValue([twoFactor, mocks.fetchState, false, undefined]);

        renderPortal();

        await waitFor(() => expect(mocks.fetchState).toHaveBeenCalled());
        expect(checkSafeRedirection).not.toHaveBeenCalled();
    });
});

describe("first factor callbacks", () => {
    const unauthenticated = { authentication_level: 0, factor_knowledge: false, username: "test" };

    function renderUnauthenticated() {
        vi.mocked(useAutheliaState).mockReturnValue([unauthenticated, mocks.fetchState, false, undefined]);

        return render(
            <MemoryRouter initialEntries={["/"]}>
                <LoginPortal {...defaultProps} />
            </MemoryRouter>,
        );
    }

    it("renders the first factor form once the state resolves", async () => {
        renderUnauthenticated();
        expect(await screen.findByTestId("first-factor-form")).toBeInTheDocument();
    });

    it("enables the form once the state resolves", async () => {
        renderUnauthenticated();

        await waitFor(() => expect(screen.getByTestId("first-factor-form")).toHaveAttribute("data-disabled", "false"));
    });

    it("disables the form while authentication is in progress", async () => {
        renderUnauthenticated();
        await screen.findByTestId("first-factor-form");

        fireEvent.click(screen.getByTestId("ff-start"));

        await waitFor(() => expect(screen.getByTestId("first-factor-form")).toHaveAttribute("data-disabled", "true"));

        fireEvent.click(screen.getByTestId("ff-stop"));

        await waitFor(() => expect(screen.getByTestId("first-factor-form")).toHaveAttribute("data-disabled", "false"));
    });

    it("redirects on a successful sign in that carries a redirect", async () => {
        renderUnauthenticated();
        await screen.findByTestId("first-factor-form");

        fireEvent.click(screen.getByTestId("ff-success-redirect"));

        await waitFor(() => expect(mocks.redirector).toHaveBeenCalledWith("https://example.com"));
    });

    it("refetches the state on a successful sign in without a redirect", async () => {
        renderUnauthenticated();
        await screen.findByTestId("first-factor-form");

        mocks.fetchState.mockClear();
        fireEvent.click(screen.getByTestId("ff-success-plain"));

        await waitFor(() => expect(mocks.fetchState).toHaveBeenCalled());
        expect(mocks.redirector).not.toHaveBeenCalled();
    });

    it("refetches the state when another tab signs in", async () => {
        renderUnauthenticated();
        await screen.findByTestId("first-factor-form");

        mocks.fetchState.mockClear();
        fireEvent.click(screen.getByTestId("ff-channel"));

        await waitFor(() => expect(mocks.fetchState).toHaveBeenCalled());
    });
});

describe("second factor route", () => {
    const oneFactor = { authentication_level: 1, factor_knowledge: true, username: "test" };
    const userInfo = {
        display_name: "test",
        emails: [],
        has_duo: false,
        has_totp: true,
        has_webauthn: false,
        method: 1,
    };

    function renderSecondFactor() {
        vi.mocked(useAutheliaState).mockReturnValue([oneFactor, mocks.fetchState, false, undefined]);
        vi.mocked(useConfiguration).mockReturnValue([
            { available_methods: new Set([1]), password_change_disabled: false, password_reset_disabled: false },
            vi.fn(),
            false,
            undefined,
        ]);
        vi.mocked(useUserInfoPOST).mockReturnValue([userInfo, mocks.fetchUserInfo, false, undefined]);

        return render(
            <MemoryRouter initialEntries={["/2fa/totp"]}>
                <LoginPortal {...defaultProps} />
            </MemoryRouter>,
        );
    }

    it("renders the second factor form", async () => {
        renderSecondFactor();
        expect(await screen.findByTestId("second-factor-form")).toHaveAttribute("data-level", "1");
    });

    it("refetches the user info when the method changes", async () => {
        renderSecondFactor();
        await screen.findByTestId("second-factor-form");

        mocks.fetchUserInfo.mockClear();
        fireEvent.click(screen.getByTestId("sf-method-changed"));

        await waitFor(() => expect(mocks.fetchUserInfo).toHaveBeenCalled());
    });

    it("redirects on a successful second factor with a redirect", async () => {
        renderSecondFactor();
        await screen.findByTestId("second-factor-form");

        fireEvent.click(screen.getByTestId("sf-success-redirect"));

        await waitFor(() => expect(mocks.redirector).toHaveBeenCalledWith("https://example.com"));
    });

    it("refetches the state on a successful second factor without a redirect", async () => {
        renderSecondFactor();
        await screen.findByTestId("second-factor-form");

        mocks.fetchState.mockClear();
        fireEvent.click(screen.getByTestId("sf-success-plain"));

        await waitFor(() => expect(mocks.fetchState).toHaveBeenCalled());
    });

    it("renders nothing without the user info", async () => {
        vi.mocked(useAutheliaState).mockReturnValue([oneFactor, mocks.fetchState, false, undefined]);
        vi.mocked(useConfiguration).mockReturnValue([
            { available_methods: new Set([1]), password_change_disabled: false, password_reset_disabled: false },
            vi.fn(),
            false,
            undefined,
        ]);

        render(
            <MemoryRouter initialEntries={["/2fa/totp"]}>
                <LoginPortal {...defaultProps} />
            </MemoryRouter>,
        );

        // Flush the lazy route import so the absence assertion cannot pass vacuously.
        await act(async () => {});

        expect(screen.queryByTestId("second-factor-form")).not.toBeInTheDocument();
    });
});

describe("authenticated route", () => {
    const userInfo = {
        display_name: "test",
        emails: [],
        has_duo: false,
        has_totp: true,
        has_webauthn: false,
        method: 1,
    };

    it("renders the authenticated view", async () => {
        vi.mocked(useAutheliaState).mockReturnValue([
            { authentication_level: 1, factor_knowledge: true, username: "test" },
            mocks.fetchState,
            false,
            undefined,
        ]);
        vi.mocked(useConfiguration).mockReturnValue([
            { available_methods: new Set(), password_change_disabled: false, password_reset_disabled: false },
            vi.fn(),
            false,
            undefined,
        ]);
        vi.mocked(useUserInfoPOST).mockReturnValue([userInfo, mocks.fetchUserInfo, false, undefined]);

        render(
            <MemoryRouter initialEntries={["/authenticated"]}>
                <LoginPortal {...defaultProps} />
            </MemoryRouter>,
        );

        expect(await screen.findByTestId("authenticated-view")).toBeInTheDocument();
    });

    it("renders nothing without the user info", async () => {
        vi.mocked(useAutheliaState).mockReturnValue([
            { authentication_level: 1, factor_knowledge: true, username: "test" },
            mocks.fetchState,
            false,
            undefined,
        ]);

        render(
            <MemoryRouter initialEntries={["/authenticated"]}>
                <LoginPortal {...defaultProps} />
            </MemoryRouter>,
        );

        // Flush the lazy route import so the absence assertion cannot pass vacuously.
        await act(async () => {});

        expect(screen.queryByTestId("authenticated-view")).not.toBeInTheDocument();
    });
});

describe("external identity notifications and link page", () => {
    const unauthenticated = { authentication_level: 0, factor_knowledge: false, username: "" };
    const twoFactor = { authentication_level: 2, factor_knowledge: true, username: "test" };

    function renderPortal(route = "/") {
        return render(
            <MemoryRouter initialEntries={[route]}>
                <LoginPortal {...defaultProps} externalIdentityLogin={true} />
            </MemoryRouter>,
        );
    }

    it("notifies a generic error when the external sign in was rejected", async () => {
        mockQueryParams = { external_identity_error: "true" };
        vi.mocked(useAutheliaState).mockReturnValue([unauthenticated, mocks.fetchState, false, undefined]);

        renderPortal();

        await waitFor(() =>
            expect(mockCreateErrorNotification).toHaveBeenCalledWith(
                "There was an issue signing in with the external provider",
            ),
        );
        expect(await screen.findByTestId("first-factor-form")).toHaveAttribute("data-openid-connect-login", "true");
        expect(mockCreateInfoNotification).not.toHaveBeenCalled();
    });

    it("does not notify an error without the parameter", async () => {
        vi.mocked(useAutheliaState).mockReturnValue([unauthenticated, mocks.fetchState, false, undefined]);

        renderPortal();

        expect(await screen.findByTestId("first-factor-form")).toHaveAttribute("data-openid-connect-login", "true");
        expect(mockCreateErrorNotification).not.toHaveBeenCalled();
        expect(mockCreateInfoNotification).not.toHaveBeenCalled();
    });

    it("shows the link page for the provider without a notification", async () => {
        mockQueryParams = { link_provider: "google" };
        vi.mocked(useAutheliaState).mockReturnValue([unauthenticated, mocks.fetchState, false, undefined]);

        renderPortal("/external-identity/link");

        const form = await screen.findByTestId("first-factor-form");

        expect(form).toHaveAttribute("data-external-identity-link", "google");
        expect(mockNavigate).not.toHaveBeenCalled();
        expect(mockPostOpenIDConnectStart).not.toHaveBeenCalled();
        expect(mockCreateInfoNotification).not.toHaveBeenCalled();
        expect(mockCreateErrorNotification).not.toHaveBeenCalled();
    });

    it("shows the plain login page when the link parameter is on the index page", async () => {
        mockQueryParams = { link_provider: "google" };
        vi.mocked(useAutheliaState).mockReturnValue([unauthenticated, mocks.fetchState, false, undefined]);

        renderPortal();

        const form = await screen.findByTestId("first-factor-form");

        expect(form).toHaveAttribute("data-external-identity-link", "");
        expect(form).toHaveAttribute("data-openid-connect-login", "true");
        expect(mockCreateInfoNotification).not.toHaveBeenCalled();
    });

    it("sends the user to the index page when the link page has no provider", async () => {
        vi.mocked(useAutheliaState).mockReturnValue([unauthenticated, mocks.fetchState, false, undefined]);

        renderPortal("/external-identity/link");

        await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith("/"));
    });

    it("continues the link when a signed in user reaches the link page", async () => {
        mockQueryParams = { link_provider: "google" };
        mockPostOpenIDConnectStart.mockResolvedValue({ authorization_url: "https://op.example.com/authorize" });
        vi.mocked(useAutheliaState).mockReturnValue([twoFactor, mocks.fetchState, false, undefined]);

        renderPortal("/external-identity/link");

        await waitFor(() => expect(mocks.redirector).toHaveBeenCalledWith("https://op.example.com/authorize"));
        expect(mockCreateInfoNotification).not.toHaveBeenCalled();
    });

    it("refreshes the state rather than redirecting after signing in on the link page", async () => {
        mockQueryParams = { link_provider: "google" };
        vi.mocked(useAutheliaState).mockReturnValue([unauthenticated, mocks.fetchState, false, undefined]);

        renderPortal("/external-identity/link");

        const button = await screen.findByTestId("ff-success-redirect");

        mocks.fetchState.mockClear();

        fireEvent.click(button);

        await waitFor(() => expect(mocks.fetchState).toHaveBeenCalledTimes(1));
        expect(mocks.redirector).not.toHaveBeenCalled();
    });
});
