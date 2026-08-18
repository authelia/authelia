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
