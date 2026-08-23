import { render } from "@testing-library/react";

import LinearProgressBar from "@components/LinearProgressBar";

function track(container: HTMLElement) {
    return container.querySelector<HTMLElement>('[data-slot="progress"]');
}

it("renders without crashing", () => {
    render(<LinearProgressBar value={40} />);
});

it("defaults the height when none is given", () => {
    const { container } = render(<LinearProgressBar value={40} />);
    expect(track(container)).toHaveStyle({ height: "8px" });
});

it("renders with adjusted height", () => {
    const { container } = render(<LinearProgressBar value={40} height={2} />);
    expect(track(container)).toHaveStyle({ height: "2px" });
});

it("keeps an explicit height of zero", () => {
    const { container } = render(<LinearProgressBar value={40} height={0} />);
    expect(track(container)).toHaveStyle({ height: "0px" });
});
