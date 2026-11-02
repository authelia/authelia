// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from "@testing-library/react";

import BaseLoadingPage from "@views/LoadingPage/BaseLoadingPage";

it("renders the loading message", () => {
    render(<BaseLoadingPage message="Please wait" />);
    expect(screen.getByText("Please wait...")).toBeInTheDocument();
});

it("does not inject a stylesheet to animate logo", () => {
    const before = document.head.querySelectorAll("style").length;

    render(<BaseLoadingPage message="Loading" />);

    expect(document.head.querySelectorAll("style")).toHaveLength(before);
});
