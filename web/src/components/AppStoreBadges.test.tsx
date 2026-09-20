// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from "@testing-library/react";

import AppStoreBadges from "@components/AppStoreBadges";

it("renders without crashing", () => {
    render(<AppStoreBadges iconSize={32} appleStoreLink="http://apple" googlePlayLink="http://google" />);
});

it("renders with target blank", () => {
    render(<AppStoreBadges iconSize={32} appleStoreLink="http://apple" googlePlayLink="http://google" targetBlank />);
});

it("renders the apple store badge before the google play badge", () => {
    const { container } = render(
        <AppStoreBadges iconSize={32} appleStoreLink="http://apple" googlePlayLink="http://google" />,
    );

    expect(Array.from(container.querySelectorAll("img")).map((img) => img.getAttribute("alt"))).toEqual([
        "apple store",
        "google play",
    ]);
});

it("omits a badge without a link", () => {
    render(<AppStoreBadges iconSize={32} appleStoreLink="http://apple" />);

    expect(screen.getByAltText("apple store")).toBeInTheDocument();
    expect(screen.queryByAltText("google play")).not.toBeInTheDocument();
});

it("omits both badges without links", () => {
    render(<AppStoreBadges iconSize={32} />);

    expect(screen.queryByAltText("apple store")).not.toBeInTheDocument();
    expect(screen.queryByAltText("google play")).not.toBeInTheDocument();
});
