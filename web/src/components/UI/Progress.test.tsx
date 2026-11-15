import { render, screen } from "@testing-library/react";

import { Progress } from "@components/UI/Progress";

function indicatorTransform(container: HTMLElement) {
    return container.querySelector<HTMLElement>('[data-slot="progress-indicator"]')?.style.transform;
}

it("derives the transform from the default maximum", () => {
    const { container } = render(<Progress value={25} />);
    expect(indicatorTransform(container)).toBe("translateX(-75%)");
});

it("derives the transform from a custom maximum", () => {
    const { container } = render(<Progress value={5} max={10} />);
    expect(indicatorTransform(container)).toBe("translateX(-50%)");
});

it("treats an absent value as zero", () => {
    const { container } = render(<Progress />);
    expect(indicatorTransform(container)).toBe("translateX(-100%)");
});

it("exposes the value to assistive technology", () => {
    render(<Progress value={25} />);
    const bar = screen.getByRole("progressbar");
    expect(bar).toHaveAttribute("aria-valuenow", "25");
    expect(bar).toHaveAttribute("aria-valuemax", "100");
});

it("exposes a custom maximum to assistive technology", () => {
    render(<Progress value={5} max={10} />);
    const bar = screen.getByRole("progressbar");
    expect(bar).toHaveAttribute("aria-valuenow", "5");
    expect(bar).toHaveAttribute("aria-valuemax", "10");
});

it("falls back to the default maximum when the maximum is zero", () => {
    const { container } = render(<Progress value={25} max={0} />);
    expect(indicatorTransform(container)).toBe("translateX(-75%)");
});

it("clamps a value above the maximum to a full track", () => {
    const { container } = render(<Progress value={150} />);
    expect(indicatorTransform(container)).toBe("translateX(-0%)");
});

it("clamps a negative value to an empty track", () => {
    const { container } = render(<Progress value={-20} />);
    expect(indicatorTransform(container)).toBe("translateX(-100%)");
});

it("normalizes an invalid maximum for the bar and for assistive technology alike", () => {
    const { container } = render(<Progress value={25} max={0} />);
    expect(indicatorTransform(container)).toBe("translateX(-75%)");

    const bar = screen.getByRole("progressbar");
    expect(bar).toHaveAttribute("aria-valuemax", "100");
    expect(bar).toHaveAttribute("aria-valuenow", "25");
});

it("measures a custom minimum as the start of the range", () => {
    const { container } = render(<Progress value={50} min={50} max={100} />);
    expect(indicatorTransform(container)).toBe("translateX(-100%)");

    const bar = screen.getByRole("progressbar");
    expect(bar).toHaveAttribute("aria-valuemin", "50");
    expect(bar).toHaveAttribute("aria-valuenow", "50");
});

it("measures a value within a custom range", () => {
    const { container } = render(<Progress value={75} min={50} max={100} />);
    expect(indicatorTransform(container)).toBe("translateX(-50%)");
});

it("treats a value that is not a number as indeterminate", () => {
    const { container } = render(<Progress value={NaN} />);
    expect(indicatorTransform(container)).toBe("translateX(-100%)");
});
