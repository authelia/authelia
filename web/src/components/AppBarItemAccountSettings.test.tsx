// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import AppBarItemAccountSettings from "@components/AppBarItemAccountSettings";
import { useFlowPresent } from "@hooks/Flow";
import { useUserSessionElevation } from "@hooks/UserSessionElevation";

const mockNavigate = vi.fn();
const mockDoSignOut = vi.fn();

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: vi.fn(() => mockNavigate),
}));

vi.mock("@hooks/SignOut", () => ({
    useSignOut: vi.fn(() => mockDoSignOut),
}));

vi.mock("@hooks/Flow", () => ({
    useFlowPresent: vi.fn(),
}));

vi.mock("@hooks/UserSessionElevation", () => ({
    useUserSessionElevation: vi.fn(),
}));

const mockFetchElevation = vi.fn();

const elevation = (requireSecondFactor: boolean) => ({
    can_skip_second_factor: false,
    elevated: false,
    expires: 0,
    factor_knowledge: true,
    require_second_factor: requireSecondFactor,
    skip_second_factor: false,
});

beforeEach(() => {
    mockFetchElevation.mockReset();
    vi.mocked(useUserSessionElevation).mockReturnValue([elevation(false), mockFetchElevation, false, undefined]);
});

const mockUserInfo = {
    display_name: "john",
    emails: ["john@example.com"],
    has_duo: false,
    has_totp: false,
    has_webauthn: false,
    method: 1,
};

const openDropdown = (button: HTMLElement) => {
    fireEvent.click(button, { button: 0 });
};

it("renders without crashing", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
});

it("does not render when no userinfo", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    const { container } = render(<AppBarItemAccountSettings />);
    expect(container).toBeEmptyDOMElement();
});

it("renders avatar with first letter of display name", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    expect(screen.getByText("J")).toBeInTheDocument();
});

it("opens menu on button click", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    openDropdown(button);
    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByText("Logout")).toBeInTheDocument();
});

it("renders switch user menu item when flow present", () => {
    vi.mocked(useFlowPresent).mockReturnValue(true);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    openDropdown(button);
    expect(screen.getByText("Switch User")).toBeInTheDocument();
});

it("does not render switch user menu item when flow not present", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    openDropdown(button);
    expect(screen.queryByText("Switch User")).not.toBeInTheDocument();
});

it("navigates to settings on settings click", async () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    openDropdown(button);
    const settingsItem = screen.getByText("Settings");
    fireEvent.click(settingsItem);
    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith("/settings"));
});

it("logs out on logout click", async () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    openDropdown(button);
    const logoutItem = screen.getByText("Logout");
    fireEvent.click(logoutItem);
    await waitFor(() => expect(mockDoSignOut).toHaveBeenCalledWith(false));
});

it("switches user on switch user click", async () => {
    vi.mocked(useFlowPresent).mockReturnValue(true);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    openDropdown(button);
    const switchItem = screen.getByText("Switch User");
    fireEvent.click(switchItem);
    await waitFor(() => expect(mockDoSignOut).toHaveBeenCalledWith(true));
});

it("fetches the elevation only when there is userinfo", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings />);
    expect(mockFetchElevation).not.toHaveBeenCalled();
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    expect(mockFetchElevation).toHaveBeenCalled();
});

it("fetches the elevation again when the userinfo changes", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    const { rerender } = render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    expect(mockFetchElevation).toHaveBeenCalledTimes(1);
    rerender(<AppBarItemAccountSettings userInfo={{ ...mockUserInfo }} />);
    expect(mockFetchElevation).toHaveBeenCalledTimes(2);
});

it("does not render settings menu item when a second factor is required", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    vi.mocked(useUserSessionElevation).mockReturnValue([elevation(true), mockFetchElevation, false, undefined]);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    openDropdown(button);
    expect(screen.queryByText("Settings")).not.toBeInTheDocument();
    expect(screen.getByText("Logout")).toBeInTheDocument();
});

it("does not render settings menu item until the elevation is fetched", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    vi.mocked(useUserSessionElevation).mockReturnValue([undefined, mockFetchElevation, true, undefined]);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    openDropdown(button);
    expect(screen.queryByText("Settings")).not.toBeInTheDocument();
    expect(screen.getByText("Logout")).toBeInTheDocument();
});
