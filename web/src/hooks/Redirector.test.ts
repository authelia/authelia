import { renderHook } from "@testing-library/react";

import { useRedirector } from "@hooks/Redirector";

it("redirects to the provided url", () => {
    const mockLocation = { href: "" };
    vi.stubGlobal("location", mockLocation);
    const { result } = renderHook(() => useRedirector());
    const redirect = result.current;
    redirect("https://example.com");
    expect(mockLocation.href).toBe("https://example.com");
});

it("returns stable callback", () => {
    const { rerender, result } = renderHook(() => useRedirector());
    const callback1 = result.current;
    rerender();
    const callback2 = result.current;
    expect(callback1).toBe(callback2);
});
