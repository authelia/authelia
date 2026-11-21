// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { act, renderHook } from "@testing-library/react";

import { useRemoteCall } from "@hooks/RemoteCall";

it("returns data on successful call", async () => {
    const mockFn = vi.fn().mockResolvedValue("data");
    const { result } = renderHook(() => useRemoteCall(mockFn));

    await act(async () => {
        result.current[1]();
    });

    expect(result.current[0]).toBe("data");
    expect(result.current[2]).toBe(false);
    expect(result.current[3]).toBeUndefined();
});

it("sets error on failed call", async () => {
    const mockError = new Error("test error");
    const mockFn = vi.fn().mockRejectedValue(mockError);
    vi.spyOn(console, "error").mockImplementation(() => {});

    const { result } = renderHook(() => useRemoteCall(mockFn));

    await act(async () => {
        result.current[1]();
    });

    expect(result.current[3]).toBe(mockError);

    vi.mocked(console.error).mockRestore();
});

it("initially returns undefined data and no error", () => {
    const mockFn = vi.fn().mockResolvedValue("data");
    const { result } = renderHook(() => useRemoteCall(mockFn));

    expect(result.current[0]).toBeUndefined();
    expect(result.current[2]).toBe(false);
    expect(result.current[3]).toBeUndefined();
});

it("clears inProgress after a failed call", async () => {
    const mockFn = vi.fn().mockRejectedValue(new Error("test error"));
    vi.spyOn(console, "error").mockImplementation(() => {});

    const { result } = renderHook(() => useRemoteCall(mockFn));

    await act(async () => {
        result.current[1]();
    });

    expect(result.current[2]).toBe(false);

    vi.mocked(console.error).mockRestore();
});

it("clears the previous error when a new call succeeds", async () => {
    const mockFn = vi.fn().mockRejectedValueOnce(new Error("test error")).mockResolvedValueOnce("data");
    vi.spyOn(console, "error").mockImplementation(() => {});

    const { result } = renderHook(() => useRemoteCall(mockFn));

    await act(async () => {
        result.current[1]();
    });

    expect(result.current[3]).toBeInstanceOf(Error);

    await act(async () => {
        result.current[1]();
    });

    expect(result.current[0]).toBe("data");
    expect(result.current[3]).toBeUndefined();

    vi.mocked(console.error).mockRestore();
});

it("retains the previous data while a new call is in flight by default", async () => {
    let resolve: (value: string) => void = () => {};
    const mockFn = vi
        .fn()
        .mockResolvedValueOnce("first")
        .mockReturnValueOnce(new Promise<string>((r) => (resolve = r)));

    const { result } = renderHook(() => useRemoteCall(mockFn));

    await act(async () => {
        result.current[1]();
    });

    act(() => {
        result.current[1]();
    });

    expect(result.current[0]).toBe("first");
    expect(result.current[2]).toBe(true);

    await act(async () => {
        resolve("second");
    });

    expect(result.current[0]).toBe("second");
});

it("discards the previous data while a new call is in flight when resetDataOnCall is set", async () => {
    let resolve: (value: string) => void = () => {};
    const mockFn = vi
        .fn()
        .mockResolvedValueOnce("first")
        .mockReturnValueOnce(new Promise<string>((r) => (resolve = r)));

    const { result } = renderHook(() => useRemoteCall(mockFn, { resetDataOnCall: true }));

    await act(async () => {
        result.current[1]();
    });

    expect(result.current[0]).toBe("first");

    act(() => {
        result.current[1]();
    });

    expect(result.current[0]).toBeUndefined();
    expect(result.current[2]).toBe(true);

    await act(async () => {
        resolve("second");
    });

    expect(result.current[0]).toBe("second");
});
