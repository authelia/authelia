import React from "react";

import { render, screen } from "@testing-library/react";

import ComponentWithTooltip from "@components/ComponentWithTooltip";

it("renders without crashing", () => {
    render(
        <ComponentWithTooltip render={false} title="test">
            <div>child</div>
        </ComponentWithTooltip>,
    );
});

it("renders children without tooltip when render is false", () => {
    const { container } = render(
        <ComponentWithTooltip render={false} title="test">
            <div>child</div>
        </ComponentWithTooltip>,
    );
    expect(screen.getByText("child")).toBeInTheDocument();
    const child = screen.getByText("child");
    expect(child.parentElement?.tagName).not.toBe("SPAN");
});

it("renders children with tooltip when render is true", () => {
    const { container } = render(
        <ComponentWithTooltip render={true} title="test">
            <div>child</div>
        </ComponentWithTooltip>,
    );
    expect(screen.getByText("child")).toBeInTheDocument();
    const child = screen.getByText("child");
    expect(child.parentElement?.tagName).toBe("SPAN");
});

it("passes tooltip props correctly", () => {
    render(
        <ComponentWithTooltip render={true} title="test title" placement="top">
            <span>child</span>
        </ComponentWithTooltip>,
    );
    expect(screen.getByText("child")).toBeInTheDocument();
});
