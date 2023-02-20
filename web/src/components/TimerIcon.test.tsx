import React from "react";

import { act, render } from "@testing-library/react";

import TimerIcon from "@components/TimerIcon";

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    // restoring date after each test run
    vi.useRealTimers();
});

it("renders without crashing", () => {
    render(<TimerIcon width={32} height={32} period={30} />);
});

it("renders a timer icon with updating progress for a given period", () => {
    const { container } = render(<TimerIcon width={32} height={32} period={30} />);
    const initialProgress = container
        .firstElementChild!.firstElementChild!.nextElementSibling!.nextElementSibling!.getAttribute("stroke-dasharray")!
        .split(/\s(.+)/)[0];

    act(() => {
        vi.advanceTimersByTime(3000);
    });

    const updatedProgress = container
        .firstElementChild!.firstElementChild!.nextElementSibling!.nextElementSibling!.getAttribute("stroke-dasharray")!
        .split(/\s(.+)/)[0];

    expect(Number(updatedProgress)).toBeGreaterThan(Number(initialProgress));
});
