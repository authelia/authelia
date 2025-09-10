import React from "react";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import AppBarItemAccountSettings from "@components/AppBarItemAccountSettings";
import AppBarItemLanguage from "@components/AppBarItemLanguage";
import AppBarLoginPortal from "@components/AppBarLoginPortal";

const mockOnLocaleChange = vi.fn();

vi.mock("@components/AppBarItemLanguage", () => ({
    default: vi.fn(() => <div>Language</div>),
}));

vi.mock("@components/AppBarItemAccountSettings", () => ({
    default: vi.fn(() => <div>Account</div>),
}));

const mockUserInfo = {
    display_name: "john",
    emails: ["john@example.com"],
    method: 1,
    has_webauthn: false,
    has_totp: false,
    has_duo: false,
};

const mockLanguages = [{ display: "English", locale: "en", fallbacks: [], namespaces: [] }];

it("renders without crashing", () => {
    render(<AppBarLoginPortal />);
});

it("renders language and account components", () => {
    render(
        <AppBarLoginPortal
            userInfo={mockUserInfo}
            localeCurrent="en"
            localeList={mockLanguages}
            onLocaleChange={mockOnLocaleChange}
        />,
    );
    expect(screen.getByText("Language")).toBeInTheDocument();
    expect(screen.getByText("Account")).toBeInTheDocument();
});

it("passes props to language component", () => {
    render(
        <AppBarLoginPortal
            userInfo={mockUserInfo}
            localeCurrent="en"
            localeList={mockLanguages}
            onLocaleChange={mockOnLocaleChange}
        />,
    );
    expect(AppBarItemLanguage).toHaveBeenCalledWith(
        {
            localeCurrent: "en",
            localeList: mockLanguages,
            onChange: mockOnLocaleChange,
        },
        undefined,
    );
});

it("passes props to account component", () => {
    render(
        <AppBarLoginPortal
            userInfo={mockUserInfo}
            localeCurrent="en"
            localeList={mockLanguages}
            onLocaleChange={mockOnLocaleChange}
        />,
    );
    expect(AppBarItemAccountSettings).toHaveBeenCalledWith(
        {
            userInfo: mockUserInfo,
        },
        undefined,
    );
});
