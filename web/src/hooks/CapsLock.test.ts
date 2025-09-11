import React from "react";

import { renderHook } from "@testing-library/react";
import { vi } from "vitest";

import { useCheckCapsLock } from "@hooks/CapsLock";

it("notifies when caps lock is on", () => {
    const mockSetCapsLockNotify = vi.fn();
    const { result } = renderHook(() => useCheckCapsLock(mockSetCapsLockNotify));
    const checkCapsLock = result.current;

    const mockEvent = {
        getModifierState: vi.fn((modifier) => modifier === "CapsLock"),
    } as any;

    checkCapsLock(mockEvent);

    expect(mockSetCapsLockNotify).toHaveBeenCalledWith(true);
});

it("notifies when caps lock is off", () => {
    const mockSetCapsLockNotify = vi.fn();
    const { result } = renderHook(() => useCheckCapsLock(mockSetCapsLockNotify));
    const checkCapsLock = result.current;

    const mockEvent = {
        getModifierState: vi.fn((modifier) => false),
    } as any;

    checkCapsLock(mockEvent);

    expect(mockSetCapsLockNotify).toHaveBeenCalledWith(false);
});

it("returns stable callback", () => {
    const mockSetCapsLockNotify = vi.fn();
    const { result, rerender } = renderHook(() => useCheckCapsLock(mockSetCapsLockNotify));

    const callback1 = result.current;
    rerender();
    const callback2 = result.current;

    expect(callback1).toBe(callback2);
});
