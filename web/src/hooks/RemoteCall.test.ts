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
