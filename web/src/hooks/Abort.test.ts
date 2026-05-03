import { renderHook } from "@testing-library/react";

import { useAbortSignal } from "@hooks/Abort";

it("returns a stable getter that yields an unaborted signal while mounted", () => {
    const { result } = renderHook(() => useAbortSignal());
    const getSignal = result.current;

    const first = getSignal();
    expect(first).toBeInstanceOf(AbortSignal);
    expect(first.aborted).toBe(false);

    const second = getSignal();
    expect(second).toBe(first);
});

it("aborts the signal on unmount", () => {
    const { result, unmount } = renderHook(() => useAbortSignal());
    const signal = result.current();

    expect(signal.aborted).toBe(false);
    unmount();
    expect(signal.aborted).toBe(true);
});

it("issues a fresh signal when called after the previous one aborted", () => {
    const { result, unmount } = renderHook(() => useAbortSignal());
    const getSignal = result.current;

    const first = getSignal();
    unmount();
    expect(first.aborted).toBe(true);

    const second = getSignal();
    expect(second).not.toBe(first);
    expect(second.aborted).toBe(false);
});

it("keeps the getter identity stable across re-renders", () => {
    const { rerender, result } = renderHook(() => useAbortSignal());
    const first = result.current;
    rerender();
    expect(result.current).toBe(first);
});
