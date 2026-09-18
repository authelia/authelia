import { act, fireEvent, render, screen } from "@testing-library/react";

import SettingsLayout from "@layouts/SettingsLayout";

const mockNavigate = vi.fn();

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@constants/constants", () => ({
    EncodedName: [81, 88, 86, 48, 97, 71, 86, 115, 97, 87, 69, 61],
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
    SecuritySubRoute: "/security",
    SettingsGroupManagementSubRoute: "/groups",
    SettingsRoute: "/settings",
    SettingsTwoFactorAuthenticationSubRoute: "/two-factor-authentication",
    SettingsUserManagementSubRoute: "/users",
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mockNavigate,
}));

const mocks = vi.hoisted(() => ({
    adminConfig: undefined as any,
    adminConfigError: undefined as Error | undefined,
    fetchAdminConfig: vi.fn(),
    fetchUserInfo: vi.fn(),
    userInfo: undefined as any,
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
    mocks.adminConfig = undefined;
    mocks.adminConfigError = undefined;
    mocks.userInfo = undefined;
    mocks.userInfoError = undefined;
    mocks.fetchAdminConfig.mockReset();
    mocks.fetchUserInfo.mockReset();
});

it("renders with settings title and menu button", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    expect(screen.getByLabelText("open drawer")).toBeInTheDocument();
    expect(screen.getAllByText("Settings").length).toBeGreaterThanOrEqual(1);
});

it("renders children", async () => {
    await act(async () => {
        render(
            <SettingsLayout>
                <div data-testid="child">Content</div>
            </SettingsLayout>,
        );
    });

    expect(screen.getByTestId("child")).toBeInTheDocument();
});

it("renders navigation items", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    expect(screen.getByText("Overview")).toBeInTheDocument();
    expect(screen.getByText("Security")).toBeInTheDocument();
    expect(screen.getByText("Two-Factor Authentication")).toBeInTheDocument();
    expect(screen.getAllByText("Close").length).toBeGreaterThanOrEqual(1);
});

it("sets the document title", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    expect(document.title).toContain("Settings");
});

it("navigates when a nav item is clicked", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Security"));
    });

    expect(mockNavigate).toHaveBeenCalledWith("/settings/security");
});

it("does not navigate when the selected nav item is clicked", async () => {
    Object.defineProperty(globalThis, "location", {
        configurable: true,
        value: { pathname: "/settings" },
        writable: true,
    });

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Overview"));
    });

    expect(mockNavigate).not.toHaveBeenCalled();

    Object.defineProperty(globalThis, "location", {
        configurable: true,
        value: { pathname: "/" },
        writable: true,
    });
});

const adminConfig = {
    admin_group: "admins",
    allow_admins_to_add_admins: false,
    enabled: true,
    group_management_enabled: true,
};

const userInfo = (groups: string[]) => ({
    display_name: "John",
    emails: [],
    groups,
    has_duo: false,
    has_totp: false,
    has_webauthn: false,
    method: 1,
    username: "john",
});

it("fetches the user info and the admin configuration on mount", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    expect(mocks.fetchUserInfo).toHaveBeenCalledOnce();
    expect(mocks.fetchAdminConfig).toHaveBeenCalledOnce();
});

it("shows the user and group management items to an admin", async () => {
    mocks.adminConfig = adminConfig;
    mocks.userInfo = userInfo(["admins", "dev"]);

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    expect(screen.getByText("User Management")).toBeInTheDocument();
    expect(screen.getByText("Group Management")).toBeInTheDocument();
    expect(document.getElementById("settings-menu-users")).toBeInTheDocument();
    expect(document.getElementById("settings-menu-groups")).toBeInTheDocument();
});

it("hides the management items from a user who is not in the admin group", async () => {
    mocks.adminConfig = adminConfig;
    mocks.userInfo = userInfo(["dev"]);

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    expect(screen.queryByText("User Management")).not.toBeInTheDocument();
    expect(screen.queryByText("Group Management")).not.toBeInTheDocument();
    expect(screen.getByText("Security")).toBeInTheDocument();
});

it("hides the management items when administration is disabled", async () => {
    mocks.adminConfig = { ...adminConfig, enabled: false };
    mocks.userInfo = userInfo(["admins"]);

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    expect(screen.queryByText("User Management")).not.toBeInTheDocument();
    expect(screen.queryByText("Group Management")).not.toBeInTheDocument();
});

it("hides only the group management item when group management is disabled", async () => {
    mocks.adminConfig = { ...adminConfig, group_management_enabled: false };
    mocks.userInfo = userInfo(["admins"]);

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    expect(screen.getByText("User Management")).toBeInTheDocument();
    expect(screen.queryByText("Group Management")).not.toBeInTheDocument();
});

it("hides the management items while the admin configuration is not loaded", async () => {
    mocks.adminConfig = undefined;
    mocks.userInfo = userInfo(["admins"]);

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    expect(screen.queryByText("User Management")).not.toBeInTheDocument();
});

it("hides the management items when either fetch fails", async () => {
    mocks.adminConfig = adminConfig;
    mocks.userInfo = userInfo(["admins"]);
    mocks.userInfoError = new Error("boom");

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    expect(screen.queryByText("User Management")).not.toBeInTheDocument();

    mocks.userInfoError = undefined;
    mocks.adminConfigError = new Error("boom");

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getAllByLabelText("open drawer")[1]);
    });

    expect(screen.queryByText("User Management")).not.toBeInTheDocument();
});

it("navigates to the user management route when its item is clicked", async () => {
    mocks.adminConfig = adminConfig;
    mocks.userInfo = userInfo(["admins"]);

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    await act(async () => {
        fireEvent.click(screen.getByText("User Management"));
    });

    expect(mockNavigate).toHaveBeenCalledWith("/settings/users");

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Group Management"));
    });

    expect(mockNavigate).toHaveBeenCalledWith("/settings/groups");
});
