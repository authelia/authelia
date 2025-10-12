import React from "react";

import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import AppBarItemAccountSettings from "@components/AppBarItemAccountSettings";
import { useFlowPresent } from "@hooks/Flow";

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

const mockUserInfo = {
    display_name: "john",
    emails: ["john@example.com"],
    method: 1,
    has_webauthn: false,
    has_totp: false,
    has_duo: false,
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
    fireEvent.click(button);
    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByText("Logout")).toBeInTheDocument();
});

it("renders switch user menu item when flow present", () => {
    vi.mocked(useFlowPresent).mockReturnValue(true);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(screen.getByText("Switch User")).toBeInTheDocument();
});

it("does not render switch user menu item when flow not present", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(screen.queryByText("Switch User")).toBeNull();
});

it("navigates to settings on settings click", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    const settingsItem = screen.getByText("Settings");
    fireEvent.click(settingsItem);
    expect(mockNavigate).toHaveBeenCalledWith("/settings");
});

it("logs out on logout click", () => {
    vi.mocked(useFlowPresent).mockReturnValue(false);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    const logoutItem = screen.getByText("Logout");
    fireEvent.click(logoutItem);
    expect(mockDoSignOut).toHaveBeenCalledWith(false);
});

it("switches user on switch user click", () => {
    vi.mocked(useFlowPresent).mockReturnValue(true);
    render(<AppBarItemAccountSettings userInfo={mockUserInfo} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    const switchItem = screen.getByText("Switch User");
    fireEvent.click(switchItem);
    expect(mockDoSignOut).toHaveBeenCalledWith(true);
});
