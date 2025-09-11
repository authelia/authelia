import { renderHook } from "@testing-library/react";
import { useSearchParams } from "react-router-dom";
import { vi } from "vitest";

import { useID, useToken } from "@hooks/Revoke";

vi.mock("react-router-dom", () => ({
    useSearchParams: vi.fn(),
}));

vi.mock("@constants/SearchParams", () => ({
    Identifier: "id",
    IdentityToken: "token",
}));

it("returns id when present", () => {
    const mockSearchParams = new URLSearchParams("id=123");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useID());
    expect(result.current).toBe("123");
});

it("returns undefined when id not present", () => {
    const mockSearchParams = new URLSearchParams("");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useID());
    expect(result.current).toBeUndefined();
});

it("returns token when present", () => {
    const mockSearchParams = new URLSearchParams("token=abc");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useToken());
    expect(result.current).toBe("abc");
});

it("returns undefined when token not present", () => {
    const mockSearchParams = new URLSearchParams("");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useToken());
    expect(result.current).toBeUndefined();
});
