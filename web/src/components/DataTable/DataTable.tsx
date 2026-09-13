import { type ReactNode, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";

import { ArrowDown, ArrowUp, Download } from "lucide-react";

import { ColumnVisibilityMenu, loadColumnVisibility } from "@components/DataTable/ColumnVisibilityMenu";
import { downloadCsv, toCsv } from "@components/DataTable/csv";
import { useVirtualRows } from "@components/DataTable/useVirtualRows";
import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { Pagination } from "@components/UI/Pagination";
import { Spinner } from "@components/UI/Spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@components/UI/Table";

const SEARCH_DEBOUNCE_MS = 250;
const DEFAULT_CONTAINER_HEIGHT = 400;

interface ColumnDef<T> {
    field: string;
    header: string;
    value: (row: T) => string;
    sortable?: boolean;
    hidden?: boolean;
    width?: string;
}

interface RowAction<T> {
    id: (row: T) => string;
    label: string;
    icon: ReactNode;
    onClick: (row: T) => void;
}

interface SortState {
    field: string;
    direction: "asc" | "desc";
}

interface PaginationState {
    page: number;
    pageSize: number;
    total: number;
}

interface DataTableProps<T> {
    id: string;
    columns: ColumnDef<T>[];
    rows: T[];
    getRowId: (row: T) => string;
    rowClassPrefix: string;
    rowActions?: RowAction<T>[];
    onRowDoubleClick?: (row: T) => void;
    sort: SortState;
    onSortChange: (s: SortState) => void;
    pagination: PaginationState;
    onPaginationChange: (p: Omit<PaginationState, "total">) => void;
    filter: string;
    onFilterChange: (q: string) => void;
    loading?: boolean;
    emptyText: string;
    csvFileName?: string;
    onExportAll?: () => Promise<T[]>;
}

function DataTable<T>({
    columns,
    csvFileName,
    emptyText,
    filter,
    getRowId,
    id,
    loading,
    onExportAll,
    onFilterChange,
    onPaginationChange,
    onRowDoubleClick,
    onSortChange,
    pagination,
    rowActions,
    rowClassPrefix,
    rows,
    sort,
}: Readonly<DataTableProps<T>>) {
    const [visibility, setVisibility] = useState<Record<string, boolean>>(() => loadColumnVisibility(id, columns));
    const [searchValue, setSearchValue] = useState(filter);
    const [scrollTop, setScrollTop] = useState(0);
    const [containerHeight, setContainerHeight] = useState(DEFAULT_CONTAINER_HEIGHT);
    const [exporting, setExporting] = useState(false);

    const containerRef = useRef<HTMLDivElement>(null);
    const debounceRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

    // Keep the (uncontrolled, debounced) search box in sync with an externally-reset `filter` prop
    // (e.g. a "clear filters" action) without an effect: adjust state during render when the prop we're
    // mirroring has changed since the last render, per the React-recommended derived-state pattern.
    const lastFilterRef = useRef(filter);

    if (lastFilterRef.current !== filter) {
        lastFilterRef.current = filter;
        setSearchValue(filter);
    }

    useEffect(
        () => () => {
            if (debounceRef.current) {
                clearTimeout(debounceRef.current);
            }
        },
        [],
    );

    // happy-dom/jsdom never compute real layout, so ref.clientHeight is 0 there; keep the sensible
    // DEFAULT_CONTAINER_HEIGHT fallback in that case instead of disabling virtualization outright.
    useLayoutEffect(() => {
        const el = containerRef.current;

        if (!el) {
            return;
        }

        if (el.clientHeight > 0) {
            setContainerHeight(el.clientHeight);
        }

        if (typeof ResizeObserver === "undefined") {
            return;
        }

        const observer = new ResizeObserver((entries) => {
            const entry = entries[0];

            if (entry && entry.contentRect.height > 0) {
                setContainerHeight(entry.contentRect.height);
            }
        });

        observer.observe(el);

        return () => observer.disconnect();
    }, []);

    const visibleColumns = useMemo(
        () => columns.filter((column) => visibility[column.field] ?? true),
        [columns, visibility],
    );

    const { bottomSpacerHeight, endIndex, startIndex, topSpacerHeight } = useVirtualRows({
        containerHeight,
        rowCount: rows.length,
        scrollTop,
    });

    const visibleRows = rows.slice(startIndex, endIndex);

    const hasActions = Boolean(rowActions && rowActions.length > 0);
    const columnCount = Math.max(visibleColumns.length + (hasActions ? 1 : 0), 1);

    const handleScroll = (event: React.UIEvent<HTMLDivElement>) => {
        setScrollTop(event.currentTarget.scrollTop);
    };

    const handleSearchChange = (value: string) => {
        setSearchValue(value);

        if (debounceRef.current) {
            clearTimeout(debounceRef.current);
        }

        debounceRef.current = setTimeout(() => onFilterChange(value), SEARCH_DEBOUNCE_MS);
    };

    // Sort toggle: clicking the currently-sorted column flips its direction (asc <-> desc); clicking a
    // different sortable column selects it fresh, starting at "asc".
    const handleSort = (column: ColumnDef<T>) => {
        if (column.sortable === false) {
            return;
        }

        if (sort.field === column.field) {
            onSortChange({ direction: sort.direction === "asc" ? "desc" : "asc", field: column.field });
        } else {
            onSortChange({ direction: "asc", field: column.field });
        }
    };

    const handleExport = async () => {
        if (!csvFileName) {
            return;
        }

        setExporting(true);

        try {
            const exportRows = onExportAll ? await onExportAll() : rows;

            downloadCsv(csvFileName, toCsv(columns, exportRows));
        } finally {
            setExporting(false);
        }
    };

    return (
        <div className="flex flex-col gap-2" data-slot="data-table" id={id}>
            <div className="flex flex-wrap items-center justify-between gap-2">
                <Input
                    id={`${id}-search`}
                    onChange={(event) => handleSearchChange(event.target.value)}
                    placeholder="Search"
                    value={searchValue}
                />
                <div className="flex items-center gap-2">
                    {csvFileName ? (
                        <Button
                            aria-label="Export CSV"
                            disabled={exporting}
                            id={`${id}-export`}
                            onClick={handleExport}
                            size="icon"
                            variant="outline"
                        >
                            <Download />
                        </Button>
                    ) : null}
                    <ColumnVisibilityMenu columns={columns} id={id} onChange={setVisibility} visibility={visibility} />
                </div>
            </div>
            <div className="relative">
                <div
                    className="max-h-[400px] overflow-y-auto rounded-md border"
                    onScroll={handleScroll}
                    ref={containerRef}
                >
                    <Table>
                        <TableHeader>
                            <TableRow>
                                {visibleColumns.map((column) => {
                                    const sortable = column.sortable ?? true;
                                    const isSorted = sort.field === column.field;

                                    return (
                                        <TableHead className={column.width} key={column.field}>
                                            {sortable ? (
                                                <button
                                                    className="flex items-center gap-1"
                                                    onClick={() => handleSort(column)}
                                                    type="button"
                                                >
                                                    {column.header}
                                                    {isSorted ? (
                                                        sort.direction === "asc" ? (
                                                            <ArrowUp className="size-3.5" />
                                                        ) : (
                                                            <ArrowDown className="size-3.5" />
                                                        )
                                                    ) : null}
                                                </button>
                                            ) : (
                                                column.header
                                            )}
                                        </TableHead>
                                    );
                                })}
                                {hasActions ? <TableHead className="w-px text-right">Actions</TableHead> : null}
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {rows.length === 0 ? (
                                <TableRow>
                                    <TableCell className="text-center text-muted-foreground" colSpan={columnCount}>
                                        {emptyText}
                                    </TableCell>
                                </TableRow>
                            ) : (
                                <>
                                    {topSpacerHeight > 0 ? (
                                        <tr aria-hidden="true" style={{ height: topSpacerHeight }}>
                                            <td colSpan={columnCount} />
                                        </tr>
                                    ) : null}
                                    {visibleRows.map((row) => {
                                        const rowId = getRowId(row);

                                        return (
                                            <TableRow
                                                className={`${rowClassPrefix}${rowId}`}
                                                key={rowId}
                                                onDoubleClick={() => onRowDoubleClick?.(row)}
                                            >
                                                {visibleColumns.map((column) => (
                                                    <TableCell key={column.field}>{column.value(row)}</TableCell>
                                                ))}
                                                {hasActions ? (
                                                    <TableCell className="text-right">
                                                        <div className="flex justify-end gap-1">
                                                            {rowActions!.map((action) => (
                                                                <Button
                                                                    aria-label={action.label}
                                                                    id={`${rowClassPrefix}${rowId}-${action.id(row)}`}
                                                                    key={action.id(row)}
                                                                    onClick={() => action.onClick(row)}
                                                                    size="icon-sm"
                                                                    variant="ghost"
                                                                >
                                                                    {action.icon}
                                                                </Button>
                                                            ))}
                                                        </div>
                                                    </TableCell>
                                                ) : null}
                                            </TableRow>
                                        );
                                    })}
                                    {bottomSpacerHeight > 0 ? (
                                        <tr aria-hidden="true" style={{ height: bottomSpacerHeight }}>
                                            <td colSpan={columnCount} />
                                        </tr>
                                    ) : null}
                                </>
                            )}
                        </TableBody>
                    </Table>
                </div>
                {loading ? (
                    <div className="absolute inset-0 flex items-center justify-center bg-background/60">
                        <Spinner size={32} />
                    </div>
                ) : null}
            </div>
            <Pagination
                onPageChange={(page) => onPaginationChange({ page, pageSize: pagination.pageSize })}
                onPageSizeChange={(pageSize) => onPaginationChange({ page: 1, pageSize })}
                page={pagination.page}
                pageSize={pagination.pageSize}
                total={pagination.total}
            />
        </div>
    );
}

export { DataTable };
export type { ColumnDef, DataTableProps, PaginationState, RowAction, SortState };
