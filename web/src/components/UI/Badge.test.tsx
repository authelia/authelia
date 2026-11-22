import { render, screen } from "@testing-library/react";

import { Badge } from "@components/UI/Badge";

it("renders with the badge data-slot", () => {
    render(<Badge>Default</Badge>);
    expect(screen.getByText("Default")).toHaveAttribute("data-slot", "badge");
});

it("merges a passed className", () => {
    render(<Badge className="custom-badge">Label</Badge>);
    expect(screen.getByText("Label")).toHaveClass("custom-badge");
});

it("defaults to the default variant", () => {
    render(<Badge>Default</Badge>);
    expect(screen.getByText("Default")).toHaveClass("bg-primary");
});

it.each([
    ["secondary", "bg-secondary"],
    ["destructive", "bg-destructive"],
    ["outline", "text-foreground"],
] as const)("applies the %s variant classes", (variant, expectedClass) => {
    render(<Badge variant={variant}>{variant}</Badge>);
    expect(screen.getByText(variant)).toHaveClass(expectedClass);
});

it("renders as a different element via the render prop", () => {
    render(<Badge render={<a href="#" />}>Link</Badge>);
    const el = screen.getByText("Link");
    expect(el.tagName).toBe("A");
    expect(el).toHaveAttribute("data-slot", "badge");
});
