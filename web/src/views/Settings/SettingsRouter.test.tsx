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
    SettingsGroupManagementSubRoute: "/groups",
    SettingsRoute: "/settings",
    SettingsTwoFactorAuthenticationSubRoute: "/two-factor-authentication",
    SettingsUserManagementSubRoute: "/users",
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

vi.mock("@views/Settings/UserManagement/UserManagementView", () => ({
    default: () => <div data-testid="user-management-view" />,
}));

vi.mock("@views/Settings/UserManagement/GroupManagementView", () => ({
    default: () => <div data-testid="group-management-view" />,
}));

const mocks = vi.hoisted(() => ({
    adminConfig: {
        admin_group: "admins",
        allow_admins_to_add_admins: false,
        enabled: true,
        group_management_enabled: true,
    } as any,
    adminConfigError: undefined as Error | undefined,
    fetchAdminConfig: vi.fn(),
    fetchUserInfo: vi.fn(),
    userInfo: {
        display_name: "John",
        emails: [],
        groups: ["admins"],
        has_duo: false,
        has_totp: false,
        has_webauthn: false,
        method: 1,
        username: "john",
    } as any,
    userInfoError: undefined as Error | undefined,
}));

vi.mock("@hooks/UserInfo", () => ({
    useUserInfoGET: () => [mocks.userInfo, mocks.fetchUserInfo, false, mocks.userInfoError],
}));

vi.mock("@hooks/UserManagement", () => ({
    useAdminConfigurationGET: () => [mocks.adminConfig, mocks.fetchAdminConfig, false, mocks.adminConfigError],
}));

beforeEach(() => {
    mockNavigate.mockReset();
    mocks.adminConfig = {
        admin_group: "admins",
        allow_admins_to_add_admins: false,
        enabled: true,
        group_management_enabled: true,
    };
    mocks.adminConfigError = undefined;
    mocks.userInfo = {
        display_name: "John",
        emails: [],
        groups: ["admins"],
        has_duo: false,
        has_totp: false,
        has_webauthn: false,
        method: 1,
        username: "john",
    };
    mocks.userInfoError = undefined;
    mocks.fetchAdminConfig.mockReset();
    mocks.fetchUserInfo.mockReset();
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

it("renders the user management view on its route", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/users"]}>
                <SettingsRouter />
            </MemoryRouter>,
        );
    });
    expect(screen.getByTestId("user-management-view")).toBeInTheDocument();
});

it("renders the group management view on its route", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/groups"]}>
                <SettingsRouter />
            </MemoryRouter>,
        );
    });
    expect(screen.getByTestId("group-management-view")).toBeInTheDocument();
});

it("navigates a non-admin away from the user management route", async () => {
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.mocked(useAutheliaState).mockReturnValue([
        { authentication_level: 1, factor_knowledge: true, username: "test" },
        vi.fn(),
        false,
        undefined,
    ]);
    mocks.userInfo = { ...mocks.userInfo, groups: ["dev"] };

    await act(async () => {
        render(
            <MemoryRouter initialEntries={["/users"]}>
                <SettingsRouter />
            </MemoryRouter>,
        );
    });

    expect(screen.queryByTestId("user-management-view")).not.toBeInTheDocument();
    expect(mockNavigate).toHaveBeenCalledWith("/settings");
});
