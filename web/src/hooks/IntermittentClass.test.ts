import { act, renderHook, waitFor } from "@testing-library/react";
import { vi } from "vitest";

import { useIntermittentClass } from "@hooks/IntermittentClass";

it("starts with empty class", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useIntermittentClass("active", 1000, 500, 200));
    expect(result.current).toBe("");
    vi.useRealTimers();
});

it("sets class after start millisecond", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useIntermittentClass("active", 1000, 500, 200));
    act(() => vi.advanceTimersByTime(200));
    expect(result.current).toBe("active");
    vi.useRealTimers();
});

it("sets class immediately when start millisecond is 0", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useIntermittentClass("active", 1000, 500, 0));
    act(() => vi.runAllTimers());
    expect(result.current).toBe("active");
    vi.useRealTimers();
});

it("sets class immediately when start millisecond is undefined", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useIntermittentClass("active", 1000, 500));
    act(() => vi.runAllTimers());
    expect(result.current).toBe("active");
    vi.useRealTimers();
});

it("removes class after active milliseconds", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useIntermittentClass("active", 1000, 500, 0));
    act(() => vi.runAllTimers());
    expect(result.current).toBe("active");
    act(() => vi.advanceTimersByTime(1000));
    expect(result.current).toBe("");
    vi.useRealTimers();
});

it("sets class again after inactive milliseconds", () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useIntermittentClass("active", 1000, 500, 0));
    act(() => vi.runAllTimers());
    expect(result.current).toBe("active");
    act(() => vi.advanceTimersByTime(1000));
    expect(result.current).toBe("");
    act(() => vi.advanceTimersByTime(500));
    expect(result.current).toBe("active");
    vi.useRealTimers();
});
