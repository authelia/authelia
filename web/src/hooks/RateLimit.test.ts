// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from "@testing-library/react";

import { useRateLimit } from "@hooks/RateLimit";

beforeEach(() => {
    vi.useFakeTimers();
});

afterEach(() => {
    vi.useRealTimers();
});

it("is not limited initially", () => {
    const { result } = renderHook(() => useRateLimit());

    expect(result.current.limited).toBe(false);
    expect(result.current.secondsRemaining).toBe(0);
});

it("counts down the seconds remaining", () => {
    const { result } = renderHook(() => useRateLimit());

    act(() => result.current.trigger(3));

    expect(result.current.secondsRemaining).toBe(3);

    act(() => {
        vi.advanceTimersByTime(1000);
    });

    expect(result.current.secondsRemaining).toBe(2);

    act(() => {
        vi.advanceTimersByTime(1000);
    });

    expect(result.current.secondsRemaining).toBe(1);

    act(() => {
        vi.advanceTimersByTime(1000);
    });

    expect(result.current.secondsRemaining).toBe(0);
    expect(result.current.limited).toBe(false);
});

it("rounds a partial second up", () => {
    const { result } = renderHook(() => useRateLimit());

    act(() => result.current.trigger(1.5));

    expect(result.current.secondsRemaining).toBe(2);

    act(() => {
        vi.advanceTimersByTime(1000);
    });

    expect(result.current.secondsRemaining).toBe(1);
    expect(result.current.limited).toBe(true);
});

it("is limited until the retry period elapses", () => {
    const onExpire = vi.fn();
    const { result } = renderHook(() => useRateLimit(onExpire));

    act(() => result.current.trigger(30));

    expect(result.current.limited).toBe(true);

    act(() => {
        vi.advanceTimersByTime(29999);
    });

    expect(result.current.limited).toBe(true);
    expect(onExpire).not.toHaveBeenCalled();

    act(() => {
        vi.advanceTimersByTime(1);
    });

    expect(result.current.limited).toBe(false);
    expect(onExpire).toHaveBeenCalledTimes(1);
});

it("calls onExpire once when the countdown and the retry period end together", () => {
    const onExpire = vi.fn();
    const { result } = renderHook(() => useRateLimit(onExpire));

    act(() => result.current.trigger(2));

    act(() => {
        vi.advanceTimersByTime(10000);
    });

    expect(result.current.limited).toBe(false);
    expect(onExpire).toHaveBeenCalledTimes(1);
});

it("stays limited for a retry period longer than the maximum timeout delay", () => {
    const onExpire = vi.fn();
    const { result } = renderHook(() => useRateLimit(onExpire));
    const retryAfter = 30 * 24 * 60 * 60;

    act(() => result.current.trigger(retryAfter));

    act(() => {
        vi.advanceTimersByTime(1000);
    });

    expect(result.current.limited).toBe(true);
    expect(result.current.secondsRemaining).toBe(retryAfter - 1);
    expect(onExpire).not.toHaveBeenCalled();

    vi.setSystemTime(Date.now() + (retryAfter - 3) * 1000);

    act(() => {
        vi.advanceTimersByTime(1000);
    });

    expect(result.current.limited).toBe(true);
    expect(result.current.secondsRemaining).toBe(1);
    expect(onExpire).not.toHaveBeenCalled();

    act(() => {
        vi.advanceTimersByTime(1000);
    });

    expect(result.current.limited).toBe(false);
    expect(onExpire).toHaveBeenCalledTimes(1);
});

it("restarts the retry period when triggered again", () => {
    const onExpire = vi.fn();
    const { result } = renderHook(() => useRateLimit(onExpire));

    act(() => result.current.trigger(10));

    act(() => {
        vi.advanceTimersByTime(5000);
    });

    act(() => result.current.trigger(10));

    act(() => {
        vi.advanceTimersByTime(5000);
    });

    expect(result.current.limited).toBe(true);
    expect(onExpire).not.toHaveBeenCalled();

    act(() => {
        vi.advanceTimersByTime(5000);
    });

    expect(result.current.limited).toBe(false);
    expect(onExpire).toHaveBeenCalledTimes(1);
});

it("clears the limit without calling onExpire", () => {
    const onExpire = vi.fn();
    const { result } = renderHook(() => useRateLimit(onExpire));

    act(() => result.current.trigger(30));
    act(() => result.current.clear());

    expect(result.current.limited).toBe(false);
    expect(result.current.secondsRemaining).toBe(0);

    act(() => {
        vi.advanceTimersByTime(30000);
    });

    expect(onExpire).not.toHaveBeenCalled();
});

it("calls the latest onExpire", () => {
    const first = vi.fn();
    const second = vi.fn();
    const { rerender, result } = renderHook(({ onExpire }) => useRateLimit(onExpire), {
        initialProps: { onExpire: first },
    });

    act(() => result.current.trigger(10));

    rerender({ onExpire: second });

    act(() => {
        vi.advanceTimersByTime(10000);
    });

    expect(first).not.toHaveBeenCalled();
    expect(second).toHaveBeenCalledTimes(1);
});

it("clears the timer on unmount", () => {
    const onExpire = vi.fn();
    const { result, unmount } = renderHook(() => useRateLimit(onExpire));

    act(() => result.current.trigger(30));

    unmount();

    act(() => {
        vi.advanceTimersByTime(30000);
    });

    expect(onExpire).not.toHaveBeenCalled();
    expect(vi.getTimerCount()).toBe(0);
});
