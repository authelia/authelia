// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from "@testing-library/react";

import { useSecondsUntil } from "@hooks/Countdown";

beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-11-12T10:00:00Z"));
});

afterEach(() => {
    vi.useRealTimers();
});

const now = Date.parse("2026-11-12T10:00:00Z") / 1000;

it("returns zero without an expiry", () => {
    const { result } = renderHook(() => useSecondsUntil(undefined));

    expect(result.current).toBe(0);
});

it("counts down to the expiry", () => {
    const { result } = renderHook(() => useSecondsUntil(now + 3));

    expect(result.current).toBe(3);

    act(() => {
        vi.advanceTimersByTime(1000);
    });
    expect(result.current).toBe(2);

    act(() => {
        vi.advanceTimersByTime(5000);
    });
    expect(result.current).toBe(0);
});

it("never returns a negative value", () => {
    const { result } = renderHook(() => useSecondsUntil(now - 10));

    expect(result.current).toBe(0);
});

it("restarts when the expiry changes", () => {
    const { rerender, result } = renderHook(({ expiresAt }) => useSecondsUntil(expiresAt), {
        initialProps: { expiresAt: now + 3 },
    });

    rerender({ expiresAt: now + 60 });

    expect(result.current).toBe(60);
});
