// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";

import { useAutheliaState } from "@hooks/State";
import { useUserSessionElevation } from "@hooks/UserSessionElevation";
import SettingsRouter from "@views/Settings/SettingsRouter";

const mockNavigate = vi.fn();

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@hooks/State", () => ({
    useAutheliaState: vi.fn(),
}));

vi.mock("@hooks/UserSessionElevation", () => ({
    useUserSessionElevation: vi.fn(),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mockNavigate,
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
    SecuritySubRoute: "/security",
    SettingsRoute: "/settings",
    SettingsTwoFactorAuthenticationSubRoute: "/two-factor-authentication",
}));

vi.mock("@layouts/SettingsLayout", () => ({
    default: (props: any) => <div data-testid="settings-layout">{props.children}</div>,
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

const elevation = (requireSecondFactor: boolean) => ({
    can_skip_second_factor: false,
    elevated: false,
    expires: 0,
    factor_knowledge: true,
    require_second_factor: requireSecondFactor,
    skip_second_factor: false,
});

const mockFetchElevation = vi.fn();

beforeEach(() => {
    mockNavigate.mockReset();
    mockFetchElevation.mockReset();
    vi.mocked(useUserSessionElevation).mockReturnValue([elevation(false), mockFetchElevation, false, undefined]);
});

const renderSettings = async () => {
    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/settings"]}>
                <Routes>
                    <Route path="/settings/*" element={<SettingsRouter />} />
                </Routes>
            </MemoryRouter>,
        );
    });
};

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

it("one factor state renders the settings when a second factor is not required", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    await renderSettings();
    expect(mockFetchElevation).toHaveBeenCalled();
    expect(mockNavigate).not.toHaveBeenCalled();
    expect(screen.getByTestId("settings-view")).toBeInTheDocument();
    expect(screen.getByTestId("settings-layout")).toBeInTheDocument();
});

it("one factor state calls navigate to index route when a second factor is required", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserSessionElevation).mockReturnValue([elevation(true), mockFetchElevation, false, undefined]);
    await renderSettings();
    expect(mockNavigate).toHaveBeenCalledWith("/");
    expect(screen.queryByTestId("settings-view")).not.toBeInTheDocument();
    expect(screen.queryByTestId("settings-layout")).not.toBeInTheDocument();
});

it("one factor state calls navigate to index route when the elevation can't be fetched", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserSessionElevation).mockReturnValue([undefined, mockFetchElevation, false, new Error("test")]);
    await renderSettings();
    expect(mockNavigate).toHaveBeenCalledWith("/");
    expect(screen.queryByTestId("settings-view")).not.toBeInTheDocument();
    expect(screen.queryByTestId("settings-layout")).not.toBeInTheDocument();
});

it("one factor state renders nothing until the elevation is fetched", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserSessionElevation).mockReturnValue([undefined, mockFetchElevation, true, undefined]);
    await renderSettings();
    expect(mockNavigate).not.toHaveBeenCalled();
    expect(screen.queryByTestId("settings-view")).not.toBeInTheDocument();
    expect(screen.queryByTestId("settings-layout")).not.toBeInTheDocument();
});

it("two factor state renders the settings without checking the elevation", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 2, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    vi.mocked(useUserSessionElevation).mockReturnValue([undefined, mockFetchElevation, false, undefined]);
    await renderSettings();
    expect(mockFetchElevation).not.toHaveBeenCalled();
    expect(mockNavigate).not.toHaveBeenCalled();
    expect(screen.getByTestId("settings-view")).toBeInTheDocument();
    expect(screen.getByTestId("settings-layout")).toBeInTheDocument();
});
