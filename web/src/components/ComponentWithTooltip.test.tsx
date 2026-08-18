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
            <button>child</button>
        </ComponentWithTooltip>,
    );
    const child = screen.getByRole("button", { name: "child" });
    expect(child).toBeInTheDocument();
    expect(child).toHaveAttribute("data-slot", "tooltip-trigger");
});

it("renders with placement prop", () => {
    render(
        <ComponentWithTooltip render={true} title="test title" placement="top">
            <span>child</span>
        </ComponentWithTooltip>,
    );
    expect(screen.getByText("child")).toBeInTheDocument();
});
