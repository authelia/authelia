// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { fireEvent, render, screen } from "@testing-library/react";

import ExternalIdentityIcon from "@views/LoginPortal/FirstFactor/ExternalIdentityIcon";

vi.mock("@assets/images/identity/openid.svg?react", () => ({
    default: () => <div data-testid="mark-openid" />,
}));

vi.mock("lucide-react", () => ({
    IdCardLanyard: () => <div data-testid="icon-fallback" />,
}));

it("prefers the configured logo over the bundled mark", () => {
    const { container } = render(<ExternalIdentityIcon type="discord" logoURI="https://cdn.example.com/logo.png" />);

    const img = container.querySelector("img");

    expect(img).toHaveAttribute("src", "https://cdn.example.com/logo.png");
    expect(screen.queryByTestId("mark-discord")).not.toBeInTheDocument();
});

it.each([["openid_connect", "mark-openid"]])(
    "renders the bundled mark for the %s type without a logo",
    (type, testid) => {
        render(<ExternalIdentityIcon type={type} />);

        expect(screen.getByTestId(testid)).toBeInTheDocument();
    },
);

it("falls back to the generic icon for a type without a bundled mark", () => {
    render(<ExternalIdentityIcon type="some_future_type" />);

    expect(screen.getByTestId("icon-fallback")).toBeInTheDocument();
});

it("falls back to the generic icon when the logo fails to load and no mark exists", () => {
    const { container } = render(
        <ExternalIdentityIcon type="some_future_type" logoURI="https://cdn.example.com/missing.png" />,
    );

    fireEvent.error(container.querySelector("img")!);

    expect(screen.getByTestId("icon-fallback")).toBeInTheDocument();
});
