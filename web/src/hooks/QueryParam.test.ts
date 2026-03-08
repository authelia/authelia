import { renderHook } from "@testing-library/react";
import { useSearchParams } from "react-router-dom";

import { useQueryParam } from "@hooks/QueryParam";

vi.mock("react-router-dom", () => ({
    useSearchParams: vi.fn(),
}));

it("returns value when query param is present", () => {
    const mockSearchParams = new URLSearchParams("test=hello");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useQueryParam("test"));
    expect(result.current).toBe("hello");
});

it("returns null when query param is absent", () => {
    const mockSearchParams = new URLSearchParams("");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useQueryParam("test"));
    expect(result.current).toBeNull();
});

it("returns undefined when query param is empty string", () => {
    const mockSearchParams = new URLSearchParams("test=");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useQueryParam("test"));
    expect(result.current).toBeUndefined();
});
