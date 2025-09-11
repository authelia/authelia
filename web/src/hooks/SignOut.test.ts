import { renderHook } from "@testing-library/react";
import { useSearchParams } from "react-router-dom";
import { vi } from "vitest";

import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useSignOut } from "@hooks/SignOut";

vi.mock("react-router-dom", () => ({
    useSearchParams: vi.fn(),
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: vi.fn(),
}));

vi.mock("@constants/SearchParams", () => ({
    RedirectionRestoreURL: "xrd",
    RedirectionURL: "rd",
}));

it("signs out without preserve", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("");
    (useRouterNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useSignOut());
    const signOut = result.current;

    signOut(false);

    expect(mockNavigate).toHaveBeenCalledWith("/logout", false, false, false);
});

it("signs out with preserve and no redirection", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("other=123");
    (useRouterNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useSignOut());
    const signOut = result.current;

    signOut(true);

    expect(mockNavigate).toHaveBeenCalledWith("/logout", true, true, true);
});

it("signs out with preserve and redirection", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("rd=https://example.com&other=123");
    (useRouterNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useSignOut());
    const signOut = result.current;

    signOut(true);

    const expectedSearch = new URLSearchParams("xrd=https://example.com&other=123");
    expect(mockNavigate).toHaveBeenCalledWith("/logout", true, true, true, expectedSearch);
});

it("returns stable callback", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("");
    (useRouterNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result, rerender } = renderHook(() => useSignOut());
    const callback1 = result.current;
    rerender();
    const callback2 = result.current;
    expect(callback1).toBe(callback2);
});
