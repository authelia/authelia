import { act, render } from "@testing-library/react";

import TimerIcon from "@components/TimerIcon";

beforeEach(() => {
    vi.useFakeTimers().setSystemTime(new Date(2023, 1, 1, 8));
});

afterEach(() => {
    vi.useRealTimers();
});

it("renders without crashing", () => {
    render(<TimerIcon width={32} height={32} period={30} />);
});

it("renders a timer icon with updating progress for a given period", async () => {
    const { container } = render(<TimerIcon width={32} height={32} period={30} />);
    const initialProgress =
        container.firstElementChild!.firstElementChild!.nextElementSibling!.nextElementSibling!.getAttribute(
            "stroke-dasharray",
        );
    expect(initialProgress).toBe("0 31.6");

    act(() => {
        vi.advanceTimersByTime(3000);
    });

    const updatedProgress =
        container.firstElementChild!.firstElementChild!.nextElementSibling!.nextElementSibling!.getAttribute(
            "stroke-dasharray",
        );
    expect(updatedProgress).toBe("3.16 31.6");
    expect(Number(updatedProgress!.split(/\s(.+)/)[0])).toBeGreaterThan(Number(initialProgress!.split(/\s(.+)/)[0]));
});
