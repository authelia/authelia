import { renderHook } from "@testing-library/react";

import { useVirtualRows } from "@components/DataTable/useVirtualRows";

it("renders the full range when the container height is not yet known", () => {
    const { result } = renderHook(() => useVirtualRows({ containerHeight: 0, rowCount: 500, scrollTop: 0 }));

    expect(result.current).toEqual({ bottomSpacerHeight: 0, endIndex: 500, startIndex: 0, topSpacerHeight: 0 });
});

it("windows a large row count to a small slice for a small container", () => {
    const { result } = renderHook(() =>
        useVirtualRows({ containerHeight: 400, rowCount: 500, rowHeight: 40, scrollTop: 0 }),
    );

    expect(result.current.startIndex).toBe(0);
    expect(result.current.endIndex).toBeLessThan(50);
    expect(result.current.topSpacerHeight).toBe(0);
    expect(result.current.bottomSpacerHeight).toBe((500 - result.current.endIndex) * 40);
});

it("shifts the window forward as scrollTop increases", () => {
    const { result } = renderHook(() =>
        useVirtualRows({ containerHeight: 400, overscan: 2, rowCount: 500, rowHeight: 40, scrollTop: 2000 }),
    );

    // scrollTop 2000 / rowHeight 40 = row 50, minus overscan 2 = 48
    expect(result.current.startIndex).toBe(48);
    expect(result.current.topSpacerHeight).toBe(48 * 40);
});

it("clamps the end index to the row count near the bottom", () => {
    const { result } = renderHook(() =>
        useVirtualRows({ containerHeight: 400, rowCount: 50, rowHeight: 40, scrollTop: 10000 }),
    );

    expect(result.current.endIndex).toBe(50);
    expect(result.current.bottomSpacerHeight).toBe(0);
});
