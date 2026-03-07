import { renderHook } from "@testing-library/react";

import { useIsMountedRef } from "@hooks/Mounted";

it("returns true when mounted", () => {
    const { result } = renderHook(() => useIsMountedRef());
    expect(result.current.current).toBe(true);
});

it("returns false after unmount", () => {
    const { result, unmount } = renderHook(() => useIsMountedRef());
    unmount();
    expect(result.current.current).toBe(false);
});
