import { useMemo } from "react";

const DEFAULT_ROW_HEIGHT = 40;
const DEFAULT_OVERSCAN = 5;

interface UseVirtualRowsOptions {
    rowCount: number;
    containerHeight: number;
    scrollTop: number;
    rowHeight?: number;
    overscan?: number;
}

interface VirtualRowsResult {
    startIndex: number;
    endIndex: number;
    topSpacerHeight: number;
    bottomSpacerHeight: number;
}

/**
 * Pure windowing calculation for a fixed-row-height list: given how many rows there are, how tall the
 * scroll container is, and its current scroll offset, returns the index range that should actually be
 * rendered plus the spacer heights (in px) needed above/below that range so the scrollbar length stays
 * correct.
 */
function useVirtualRows({
    containerHeight,
    overscan = DEFAULT_OVERSCAN,
    rowCount,
    rowHeight = DEFAULT_ROW_HEIGHT,
    scrollTop,
}: UseVirtualRowsOptions): VirtualRowsResult {
    return useMemo(() => {
        if (rowCount <= 0 || containerHeight <= 0 || rowHeight <= 0) {
            return { bottomSpacerHeight: 0, endIndex: rowCount, startIndex: 0, topSpacerHeight: 0 };
        }

        const visibleCount = Math.ceil(containerHeight / rowHeight);
        const startIndex = Math.max(0, Math.floor(scrollTop / rowHeight) - overscan);
        const endIndex = Math.min(rowCount, startIndex + visibleCount + overscan * 2);

        return {
            bottomSpacerHeight: (rowCount - endIndex) * rowHeight,
            endIndex,
            startIndex,
            topSpacerHeight: startIndex * rowHeight,
        };
    }, [containerHeight, overscan, rowCount, rowHeight, scrollTop]);
}

export { useVirtualRows, DEFAULT_ROW_HEIGHT, DEFAULT_OVERSCAN };
export type { UseVirtualRowsOptions, VirtualRowsResult };
