import { type ReactNode, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";

import {
    type OnChangeFn,
    type RowData,
    type SortingState,
    type ColumnDef as TanStackColumnDef,
    type Updater,
    flexRender,
    useTable,
} from "@tanstack/react-table";
import { Download } from "lucide-react";
import { useTranslation } from "react-i18next";

import { ColumnHeaderCell } from "@components/DataTable/ColumnHeaderCell";
import { downloadCsv, toCsv } from "@components/DataTable/csv";
import { type ColumnKind, dataTableFeatures } from "@components/DataTable/features";
import { ManageColumnsDialog } from "@components/DataTable/ManageColumnsDialog";
import { TruncatedCellValue } from "@components/DataTable/TruncatedCellValue";
import { type PersistedTableState, useTableState } from "@components/DataTable/useTableState";
import { useVirtualRows } from "@components/DataTable/useVirtualRows";
import { Button } from "@components/UI/Button";
import { Empty, EmptyContent, EmptyTitle } from "@components/UI/Empty";
import { Input } from "@components/UI/Input";
import { Pagination } from "@components/UI/Pagination";
import { Spinner } from "@components/UI/Spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@components/UI/Table";
import { cn } from "@utils/Styles";

const SEARCH_DEBOUNCE_MS = 250;
const DEFAULT_CONTAINER_HEIGHT = 400;
const DEFAULT_PAGE_SIZE = 25;
const DEFAULT_COLUMN_WIDTH = 200;
const ACTIONS_COLUMN_ID = "__actions__";
const HEADER_CHROME_WIDTH = 70;
const MIN_COLUMN_WIDTH = 60;
let headerMeasureContext: CanvasRenderingContext2D | null | undefined;

function measureHeaderTextWidth(text: string): number {
    if (headerMeasureContext === undefined) {
        headerMeasureContext =
            typeof document === "undefined" ? null : document.createElement("canvas").getContext("2d");
    }

    if (!headerMeasureContext) {
        return text.length * 7;
    }

    headerMeasureContext.font = "500 14px Roboto, Helvetica, Arial, sans-serif";

    return headerMeasureContext.measureText(text).width;
}

interface ColumnDef<T extends RowData> {
    field: string;
    filterable?: boolean;
    header: string;
    hidden?: boolean;
    kind?: ColumnKind;
    raw?: (row: T) => null | number;
    resizable?: boolean;
    size?: number;
    sortable?: boolean;
    value: (row: T) => string;
}

interface RowAction<T extends RowData> {
    destructive?: boolean;
    id: (row: T) => string;
    label: string;
    icon: ReactNode;
    onClick: (row: T) => void;
}

interface DataTableProps<T extends RowData> {
    columns: ColumnDef<T>[];
    csvFileName?: string;
    emptyText: string;
    getRowId: (row: T) => string;
    id: string;
    initialPageSize?: number;
    initialSort?: { direction: "asc" | "desc"; field: string };
    loading?: boolean;
    onExportAll?: () => Promise<T[]>;
    onRowDoubleClick?: (row: T) => void;
    rowActions?: RowAction<T>[];
    rowClassPrefix: string;
    rows: T[];
    toolbarEnd?: ReactNode;
}

function applyUpdater<S>(updater: Updater<S>, previous: S): S {
    return typeof updater === "function" ? (updater as (old: S) => S)(previous) : updater;
}

function DataTable<T extends RowData>({
    columns,
    csvFileName,
    emptyText,
    getRowId,
    id,
    initialPageSize = DEFAULT_PAGE_SIZE,
    initialSort,
    loading,
    onExportAll,
    onRowDoubleClick,
    rowActions,
    rowClassPrefix,
    rows,
    toolbarEnd,
}: Readonly<DataTableProps<T>>) {
    const { t: translate } = useTranslation("settings");

    const initialSorting = useMemo<SortingState>(
        () => (initialSort ? [{ desc: initialSort.direction === "desc", id: initialSort.field }] : []),
        [initialSort],
    );

    const {
        reset: resetFilters,
        state: persisted,
        update,
    } = useTableState(id, columns, initialSorting, initialPageSize);
    const [pageIndex, setPageIndex] = useState(0);
    const [manageColumnsOpen, setManageColumnsOpen] = useState(false);
    const [searchValue, setSearchValue] = useState(persisted.globalFilter);
    const [scrollTop, setScrollTop] = useState(0);
    const [containerHeight, setContainerHeight] = useState(DEFAULT_CONTAINER_HEIGHT);
    const [exporting, setExporting] = useState(false);

    const containerRef = useRef<HTMLDivElement>(null);
    const debounceRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

    const lastGlobalFilterRef = useRef(persisted.globalFilter);

    if (lastGlobalFilterRef.current !== persisted.globalFilter) {
        lastGlobalFilterRef.current = persisted.globalFilter;
        setSearchValue(persisted.globalFilter);
    }

    const lastFiltersRef = useRef({ columnFilters: persisted.columnFilters, globalFilter: persisted.globalFilter });

    if (
        lastFiltersRef.current.columnFilters !== persisted.columnFilters ||
        lastFiltersRef.current.globalFilter !== persisted.globalFilter
    ) {
        lastFiltersRef.current = { columnFilters: persisted.columnFilters, globalFilter: persisted.globalFilter };
        setPageIndex(0);
    }

    useEffect(
        () => () => {
            if (debounceRef.current) {
                clearTimeout(debounceRef.current);
            }
        },
        [],
    );

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

    const hasActions = Boolean(rowActions && rowActions.length > 0);
    const actionsLabel = translate("Actions");

    const getRowIdRef = useRef(getRowId);

    getRowIdRef.current = getRowId;

    const rowActionsRef = useRef(rowActions);

    rowActionsRef.current = rowActions;

    const tanstackColumns = useMemo(() => {
        const dataColumns = columns.map((column): TanStackColumnDef<typeof dataTableFeatures, T> => {
            const kind = column.kind ?? "string";
            const minSize = Math.max(
                MIN_COLUMN_WIDTH,
                Math.ceil(measureHeaderTextWidth(column.header)) + HEADER_CHROME_WIDTH,
            );

            return {
                accessorFn: (row) => (kind === "string" ? column.value(row) : (column.raw?.(row) ?? null)),
                cell: ({ row }) => <TruncatedCellValue value={column.value(row.original)} />,
                enableColumnFilter: column.filterable ?? true,
                enableHiding: true,
                enableResizing: column.resizable ?? true,
                enableSorting: column.sortable ?? true,
                filterFn: "columnFilter",
                header: column.header,
                id: column.field,
                meta: { display: (row: T) => column.value(row), kind },
                minSize,
                size: Math.max(column.size ?? DEFAULT_COLUMN_WIDTH, minSize),
                sortFn: kind === "string" ? "alphanumeric" : "basic",
            };
        });

        if (!hasActions) {
            return dataColumns;
        }

        const actionsColumn: TanStackColumnDef<typeof dataTableFeatures, T> = {
            cell: ({ row }) => (
                <div className="flex justify-end gap-1">
                    {rowActionsRef.current!.map((action) => (
                        <Button
                            aria-label={action.label}
                            color={action.destructive ? "destructive" : "default"}
                            id={`${rowClassPrefix}${getRowIdRef.current(row.original)}-${action.id(row.original)}`}
                            key={action.id(row.original)}
                            onClick={() => action.onClick(row.original)}
                            size="icon-sm"
                            variant="ghost"
                        >
                            {action.icon}
                        </Button>
                    ))}
                </div>
            ),
            enableColumnFilter: false,
            enableGlobalFilter: false,
            enableHiding: false,
            enableResizing: false,
            enableSorting: false,
            header: actionsLabel,
            id: ACTIONS_COLUMN_ID,
            size: rowActionsRef.current!.length * 32 + 16,
        };

        return [...dataColumns, actionsColumn];
    }, [actionsLabel, columns, hasActions, rowClassPrefix]);

    const makeStateChange = <K extends keyof PersistedTableState>(key: K): OnChangeFn<PersistedTableState[K]> => {
        return (updater) => {
            update((previous) => ({ ...previous, [key]: applyUpdater(updater, previous[key]) }));
        };
    };

    const table = useTable({
        columnResizeMode: "onChange",
        columns: tanstackColumns,
        data: rows,
        enableSortingRemoval: true,
        features: dataTableFeatures,
        getRowId: (row) => getRowId(row),
        globalFilterFn: "globalSearch",
        onColumnFiltersChange: makeStateChange("columnFilters"),
        onColumnOrderChange: makeStateChange("columnOrder"),
        onColumnSizingChange: makeStateChange("columnSizing"),
        onColumnVisibilityChange: makeStateChange("columnVisibility"),
        onGlobalFilterChange: makeStateChange("globalFilter"),
        onPaginationChange: (updater) => {
            const next = applyUpdater(updater, { pageIndex, pageSize: persisted.pageSize });

            setPageIndex(next.pageIndex);

            if (next.pageSize !== persisted.pageSize) {
                update((previous) => ({ ...previous, pageSize: next.pageSize }));
            }
        },
        onSortingChange: makeStateChange("sorting"),
        state: {
            columnFilters: persisted.columnFilters,
            columnOrder: persisted.columnOrder,
            columnSizing: persisted.columnSizing,
            columnVisibility: persisted.columnVisibility,
            globalFilter: persisted.globalFilter,
            pagination: { pageIndex, pageSize: persisted.pageSize },
            sorting: persisted.sorting,
        },
    });

    const pageRows = table.getRowModel().rows;

    const { bottomSpacerHeight, endIndex, startIndex, topSpacerHeight } = useVirtualRows({
        containerHeight,
        rowCount: pageRows.length,
        scrollTop,
    });

    const visibleRows = pageRows.slice(startIndex, endIndex);
    const columnCount = Math.max(table.getVisibleLeafColumns().length, 1);
    const hasActiveFilters = persisted.columnFilters.length > 0 || persisted.globalFilter.trim() !== "";
    const columnLabels = useMemo(
        () => Object.fromEntries(columns.map((column) => [column.field, column.header])),
        [columns],
    );

    const handleScroll = (event: React.UIEvent<HTMLDivElement>) => {
        setScrollTop(event.currentTarget.scrollTop);
    };

    const handleSearchChange = (value: string) => {
        setSearchValue(value);

        if (debounceRef.current) {
            clearTimeout(debounceRef.current);
        }

        debounceRef.current = setTimeout(() => table.setGlobalFilter(value), SEARCH_DEBOUNCE_MS);
    };

    const handleExport = async () => {
        if (!csvFileName) {
            return;
        }

        setExporting(true);

        try {
            const exportRows = onExportAll
                ? await onExportAll()
                : table.getFilteredRowModel().rows.map((row) => row.original);

            downloadCsv(csvFileName, toCsv(columns, exportRows));
        } finally {
            setExporting(false);
        }
    };

    return (
        <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-1" data-slot="data-table" id={id}>
            <div className="flex min-w-0 shrink-0 items-center justify-between gap-2">
                <Input
                    className="h-10 min-w-0 flex-1 sm:w-[30%] sm:flex-none sm:min-w-[200px]"
                    id={`${id}-search`}
                    onChange={(event) => handleSearchChange(event.target.value)}
                    placeholder={translate("Search")}
                    value={searchValue}
                />
                <div className="flex shrink-0 items-center gap-2">
                    {csvFileName ? (
                        <Button
                            aria-label={translate("Export CSV")}
                            disabled={exporting}
                            id={`${id}-export`}
                            onClick={handleExport}
                            size="icon"
                            variant="outline"
                        >
                            <Download />
                        </Button>
                    ) : null}
                    {toolbarEnd}
                </div>
            </div>
            <div className="relative min-h-0 min-w-0 flex-1">
                <div
                    className="h-full overflow-y-auto rounded-md border"
                    data-slot="table-container"
                    onScroll={handleScroll}
                    ref={containerRef}
                >
                    <Table className="table-fixed" style={{ minWidth: table.getTotalSize() }}>
                        <colgroup>
                            {table.getVisibleLeafColumns().map((column) => (
                                <col key={column.id} style={{ width: column.getSize() }} />
                            ))}
                        </colgroup>
                        <TableHeader>
                            <TableRow>
                                {table.getHeaderGroups()[0]?.headers.map((header, index, headers) => {
                                    const isActions = header.column.id === ACTIONS_COLUMN_ID;
                                    const columnDef = columns.find((candidate) => candidate.field === header.column.id);
                                    const isLast = index === headers.length - 1;

                                    return (
                                        <TableHead
                                            className={cn("relative", isActions && "text-right", !isLast && "border-r")}
                                            key={header.id}
                                        >
                                            {isActions ? (
                                                translate("Actions")
                                            ) : (
                                                <ColumnHeaderCell
                                                    columnOrder={persisted.columnOrder}
                                                    header={header}
                                                    kind={columnDef?.kind ?? "string"}
                                                    label={columnDef?.header ?? header.column.id}
                                                    onOpenManageColumns={() => setManageColumnsOpen(true)}
                                                    table={table}
                                                />
                                            )}
                                        </TableHead>
                                    );
                                })}
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {pageRows.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={columnCount}>
                                        {hasActiveFilters ? (
                                            <Empty className="border-none p-4">
                                                <EmptyTitle>
                                                    {translate("There are no results for the current filter(s).")}
                                                </EmptyTitle>
                                                <EmptyContent>
                                                    <Button onClick={resetFilters} size="sm" variant="outline">
                                                        {translate("Clear Filters")}
                                                    </Button>
                                                </EmptyContent>
                                            </Empty>
                                        ) : (
                                            <div className="text-center text-muted-foreground">{emptyText}</div>
                                        )}
                                    </TableCell>
                                </TableRow>
                            ) : (
                                <>
                                    {topSpacerHeight > 0 ? (
                                        <tr aria-hidden="true" style={{ height: topSpacerHeight }}>
                                            <td colSpan={columnCount} />
                                        </tr>
                                    ) : null}
                                    {visibleRows.map((row) => (
                                        <TableRow
                                            className={`${rowClassPrefix}${row.id}`}
                                            key={row.id}
                                            onDoubleClick={() => onRowDoubleClick?.(row.original)}
                                        >
                                            {row.getVisibleCells().map((cell, index, cells) => (
                                                <TableCell
                                                    className={cn(
                                                        cell.column.id === ACTIONS_COLUMN_ID && "text-right",
                                                        index !== cells.length - 1 && "border-r",
                                                    )}
                                                    key={cell.id}
                                                >
                                                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                                </TableCell>
                                            ))}
                                        </TableRow>
                                    ))}
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
            <div className="w-full min-w-0 shrink-0">
                <Pagination
                    onPageChange={(page) => table.setPageIndex(page - 1)}
                    onPageSizeChange={(pageSize) => table.setPageSize(pageSize)}
                    page={pageIndex + 1}
                    pageSize={persisted.pageSize}
                    total={table.getFilteredRowModel().rows.length}
                />
            </div>
            <ManageColumnsDialog
                columnOrder={persisted.columnOrder}
                id={`${id}-manage-columns`}
                labels={columnLabels}
                onOpenChange={setManageColumnsOpen}
                open={manageColumnsOpen}
                table={table}
            />
        </div>
    );
}

export { DataTable };
export type { ColumnDef, DataTableProps, RowAction };
