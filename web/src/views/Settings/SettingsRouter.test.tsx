import { act, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";

import { useAutheliaState } from "@hooks/State";
import SettingsRouter from "@views/Settings/SettingsRouter";

const mockNavigate = vi.fn();

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/State", () => ({
    useAutheliaState: vi.fn(),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mockNavigate,
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
    SecuritySubRoute: "/security",
    SettingsOpenIDConnectSubRoute: "/openid-connect",
    SettingsRoute: "/settings",
    SettingsTwoFactorAuthenticationSubRoute: "/two-factor-authentication",
}));

vi.mock("@layouts/SettingsLayout", () => ({
    default: (props: any) => <div>{props.children}</div>,
}));

vi.mock("@views/Settings/SettingsView", () => ({
    default: () => <div data-testid="settings-view" />,
}));

vi.mock("@views/Settings/Security/SecurityView", () => ({
    default: () => <div data-testid="security-view" />,
}));

vi.mock("@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationView", () => ({
    default: () => <div data-testid="2fa-view" />,
}));

vi.mock("@views/Settings/OpenIDConnect/OpenIDConnectView", () => ({
    default: () => <div data-testid="openid-connect-view" />,
}));

beforeEach(() => {
    mockNavigate.mockReset();
    document.body.dataset.openidconnectlogin = "false";
});

afterEach(() => {
    document.body.dataset.openidconnectlogin = "false";
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders without crashing", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/settings"]}>
                <SettingsRouter />
            </MemoryRouter>,
        );
    });
});

it("unauthenticated state calls navigate to index route", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 0, factor_knowledge: false, username: "" },
        vi.fn(),
        false,
        undefined,
    ]);
    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/settings"]}>
                <SettingsRouter />
            </MemoryRouter>,
        );
    });
    expect(mockNavigate).toHaveBeenCalledWith("/");
});

it("fetchStateError calls navigate to index route", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([undefined, vi.fn(), false, new Error("test")]);
    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/settings"]}>
                <SettingsRouter />
            </MemoryRouter>,
        );
    });
    expect(mockNavigate).toHaveBeenCalledWith("/");
});

it("authenticated state does not call navigate", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/settings"]}>
                <SettingsRouter />
            </MemoryRouter>,
        );
    });
    expect(mockNavigate).not.toHaveBeenCalled();
});

it("does not register the linked accounts route when no provider is configured", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/openid-connect"]}>
                <SettingsRouter />
            </MemoryRouter>,
        );
    });

    expect(screen.queryByTestId("openid-connect-view")).not.toBeInTheDocument();
});

it("registers the linked accounts route when a provider is configured", async () => {
    document.body.dataset.openidconnectlogin = "true";

    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/openid-connect"]}>
                <SettingsRouter />
            </MemoryRouter>,
        );
    });

    expect(screen.getByTestId("openid-connect-view")).toBeInTheDocument();
});
