import { render, screen } from "@testing-library/react";

import ComponentWithTooltip from "@components/ComponentWithTooltip";

beforeEach(() => {
    vi.spyOn(console, "error").mockImplementation(() => {});
});

it("renders without crashing", () => {
    render(
        <ComponentWithTooltip render={false} title="test">
            <div>child</div>
        </ComponentWithTooltip>,
    );
});

it("renders children without tooltip when render is false", () => {
    render(
        <ComponentWithTooltip render={false} title="test">
            <div>child</div>
        </ComponentWithTooltip>,
    );
    expect(screen.getByText("child")).toBeInTheDocument();
});

it("renders children with tooltip when render is true", () => {
    render(
        <ComponentWithTooltip render={true} title="test">
            <div>child</div>
        </ComponentWithTooltip>,
    );
    expect(screen.getByText("child")).toBeInTheDocument();
    const child = screen.getByText("child");
    expect(child.parentElement?.tagName).toBe("SPAN");
});

it("renders with placement prop", () => {
    render(
        <ComponentWithTooltip render={true} title="test title" placement="top">
            <span>child</span>
        </ComponentWithTooltip>,
    );
    expect(screen.getByText("child")).toBeInTheDocument();
});
