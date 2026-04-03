import { renderHook } from "@testing-library/react";
import { useSearchParams } from "react-router-dom";

import { useUserCode } from "@hooks/OpenIDConnect";

vi.mock("react-router-dom", () => ({
    useSearchParams: vi.fn(),
}));

vi.mock("@constants/SearchParams", () => ({
    UserCode: "user_code",
}));

it("returns user code when present", () => {
    const mockSearchParams = new URLSearchParams("user_code=abc123");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useUserCode());
    expect(result.current).toBe("abc123");
});

it("returns undefined when user code is absent", () => {
    const mockSearchParams = new URLSearchParams("");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useUserCode());
    expect(result.current).toBeUndefined();
});
