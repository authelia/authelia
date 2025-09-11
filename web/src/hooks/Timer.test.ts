import { act, renderHook } from "@testing-library/react";
import { vi } from "vitest";

import { useTimer } from "@hooks/Timer";

beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(0));
});

afterEach(() => {
    vi.useRealTimers();
});

it("initially returns 0 percent", () => {
    const { result } = renderHook(() => useTimer(1000));
    expect(result.current[0]).toBe(0);
});

it("trigger sets percent to 0 and starts timer", () => {
    const { result } = renderHook(() => useTimer(1000));
    const [, trigger] = result.current;
    act(() => trigger());
    expect(result.current[0]).toBe(0);
});

it("updates percent over time", () => {
    const { result } = renderHook(() => useTimer(1000));
    const [, trigger] = result.current;
    act(() => trigger());

    act(() => {
        vi.advanceTimersByTime(500);
        vi.setSystemTime(new Date(500));
    });

    expect(result.current[0]).toBeCloseTo(50, 0);
});

it("reaches 100 percent and stops", () => {
    const { result } = renderHook(() => useTimer(1000));
    const [, trigger] = result.current;
    act(() => trigger());

    act(() => {
        vi.advanceTimersByTime(1000);
        vi.setSystemTime(new Date(1000));
    });

    expect(result.current[0]).toBe(100);

    act(() => {
        vi.advanceTimersByTime(500);
        vi.setSystemTime(new Date(1500));
    });

    expect(result.current[0]).toBe(100);
});

it("clear resets percent to 0 and stops timer", () => {
    const { result } = renderHook(() => useTimer(1000));
    const [, trigger, clear] = result.current;
    act(() => trigger());

    act(() => {
        vi.advanceTimersByTime(500);
        vi.setSystemTime(new Date(500));
    });

    expect(result.current[0]).toBeGreaterThan(0);

    act(() => clear());
    expect(result.current[0]).toBe(0);

    act(() => {
        vi.advanceTimersByTime(500);
        vi.setSystemTime(new Date(1000));
    });

    expect(result.current[0]).toBe(0);
});

it("returns stable functions", () => {
    const { result, rerender } = renderHook(() => useTimer(1000));
    const [, trigger1, clear1] = result.current;
    rerender();
    const [, trigger2, clear2] = result.current;
    expect(trigger1).toBe(trigger2);
    expect(clear1).toBe(clear2);
});

it("cleans up interval on unmount", () => {
    const { result, unmount } = renderHook(() => useTimer(1000));
    const [, trigger] = result.current;
    act(() => trigger());
    unmount();
    vi.advanceTimersByTime(500);
    vi.setSystemTime(new Date(500));
    expect(result.current[0]).toBe(0);
});
