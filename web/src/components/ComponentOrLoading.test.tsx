import React from "react";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import ComponentOrLoading from "@components/ComponentOrLoading";

vi.mock("@views/LoadingPage/LoadingPage", () => ({
    default: () => <div>Loading</div>,
}));

it("renders without crashing", () => {
    render(
        <ComponentOrLoading ready={false}>
            <div>Child</div>
        </ComponentOrLoading>,
    );
});

it("renders loading component and does not render children when not ready", () => {
    const { container } = render(
        <ComponentOrLoading ready={false}>
            <div>Child</div>
        </ComponentOrLoading>,
    );

    expect(screen.getByText("Loading")).toBeInTheDocument();

    const box = container.querySelector(".MuiBox-root");
    expect(box).not.toHaveClass("hidden");

    expect(screen.queryByText("Child")).toBeNull();
});

it("renders children and hides loading component when ready", () => {
    const { container } = render(
        <ComponentOrLoading ready={true}>
            <div>Child</div>
        </ComponentOrLoading>,
    );

    expect(screen.getByText("Child")).toBeInTheDocument();

    const box = container.querySelector(".MuiBox-root");
    expect(box).toHaveClass("hidden");

    expect(screen.getByText("Loading")).toBeInTheDocument();
});
