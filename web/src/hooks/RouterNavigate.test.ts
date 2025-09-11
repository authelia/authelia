import { renderHook } from "@testing-library/react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { vi } from "vitest";

import { useRouterNavigate } from "@hooks/RouterNavigate";

vi.mock("react-router-dom", () => ({
    useNavigate: vi.fn(),
    useSearchParams: vi.fn(),
}));

it("navigates with search params override", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("");
    (useNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useRouterNavigate());
    const navigate = result.current;

    const override = new URLSearchParams("test=123");
    navigate("/path", true, true, true, override);

    expect(mockNavigate).toHaveBeenCalledWith({ pathname: "/path", search: "?test=123" });
});

it("navigates with preserve search params", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("existing=456");
    (useNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useRouterNavigate());
    const navigate = result.current;

    navigate("/path", true, true, true);

    expect(mockNavigate).toHaveBeenCalledWith({ pathname: "/path", search: "?existing=456" });
});

it("navigates with preserve flow", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams(
        "flow=login&subflow=auth&flow_id=123&user_code=456&rd=https://example.com",
    );
    (useNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useRouterNavigate());
    const navigate = result.current;

    navigate("/path", false, true, false);

    expect(mockNavigate).toHaveBeenCalledWith({
        pathname: "/path",
        search: "?flow=login&subflow=auth&flow_id=123&user_code=456",
    });
});

it("navigates with preserve redirection", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("rd=https://example.com&flow=login");
    (useNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useRouterNavigate());
    const navigate = result.current;

    navigate("/path", false, false, true);

    expect(mockNavigate).toHaveBeenCalledWith({ pathname: "/path", search: "?rd=https%3A%2F%2Fexample.com" });
});

it("navigates with no params", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("");
    (useNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useRouterNavigate());
    const navigate = result.current;

    navigate("/path");

    expect(mockNavigate).toHaveBeenCalledWith({ pathname: "/path" });
});

it("navigates with params but no preservation", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("existing=456");
    (useNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result } = renderHook(() => useRouterNavigate());
    const navigate = result.current;

    navigate("/path", false, false, false);

    expect(mockNavigate).toHaveBeenCalledWith({ pathname: "/path" });
});

it("returns stable callback", () => {
    const mockNavigate = vi.fn();
    const mockSearchParams = new URLSearchParams("");
    (useNavigate as any).mockReturnValue(mockNavigate);
    (useSearchParams as any).mockReturnValue([mockSearchParams]);

    const { result, rerender } = renderHook(() => useRouterNavigate());
    const callback1 = result.current;
    rerender();
    const callback2 = result.current;
    expect(callback1).toBe(callback2);
});
