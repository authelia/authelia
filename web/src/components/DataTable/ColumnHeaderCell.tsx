import { type DragEvent, useState } from "react";

import { type Header, type RowData, type Table } from "@tanstack/react-table";
import { ArrowDownNarrowWide, ArrowUpWideNarrow, MoreVertical } from "lucide-react";
import { useTranslation } from "react-i18next";

import { ColumnFilterPanel } from "@components/DataTable/ColumnFilterPanel";
import { type ColumnKind, type DataTableFeatures } from "@components/DataTable/features";
import { Button } from "@components/UI/Button";
import { DropdownMenuSeparator } from "@components/UI/DropdownMenu";
import { Popover, PopoverContent, PopoverTrigger } from "@components/UI/Popover";
import { cn } from "@utils/Styles";

interface ColumnHeaderCellProps<T extends RowData> {
    columnOrder: string[];
    header: Header<DataTableFeatures, T, unknown>;
    kind: ColumnKind;
    label: string;
    onOpenManageColumns: () => void;
    table: Table<DataTableFeatures, T>;
}

function reorderColumns<T extends RowData>(
    table: Table<DataTableFeatures, T>,
    columnOrder: string[],
    sourceId: string,
    targetId: string,
): void {
    const order = [...columnOrder];
    const from = order.indexOf(sourceId);
    const to = order.indexOf(targetId);

    if (from === -1 || to === -1 || from === to) {
        return;
    }

    order.splice(from, 1);
    order.splice(to, 0, sourceId);
    table.setColumnOrder(order);
}

function ColumnHeaderCell<T extends RowData>({
    columnOrder,
    header,
    kind,
    label,
    onOpenManageColumns,
    table,
}: Readonly<ColumnHeaderCellProps<T>>) {
    const { t: translate } = useTranslation("settings");
    const [open, setOpen] = useState(false);

    const column = header.column;
    const sortState = column.getIsSorted();
    const canSort = column.getCanSort();
    const canFilter = column.getCanFilter();
    const canHide = column.getCanHide();
    const isFiltered = column.getIsFiltered();

    const handleDragStart = (event: DragEvent<HTMLDivElement>) => {
        event.dataTransfer.effectAllowed = "move";
        event.dataTransfer.setData("text/plain", column.id);
    };

    const handleDragOver = (event: DragEvent<HTMLDivElement>) => {
        event.preventDefault();
        event.dataTransfer.dropEffect = "move";
    };

    const handleDrop = (event: DragEvent<HTMLDivElement>) => {
        event.preventDefault();

        const sourceId = event.dataTransfer.getData("text/plain");

        if (sourceId) {
            reorderColumns(table, columnOrder, sourceId, column.id);
        }
    };

    return (
        <>
            <div
                className="flex items-center gap-1 pr-2"
                draggable
                onDragOver={handleDragOver}
                onDragStart={handleDragStart}
                onDrop={handleDrop}
            >
                {canSort ? (
                    <button
                        className="flex items-center gap-1"
                        onClick={column.getToggleSortingHandler()}
                        type="button"
                    >
                        {label}
                        {sortState === "asc" ? (
                            <ArrowDownNarrowWide className="size-3.5" />
                        ) : sortState === "desc" ? (
                            <ArrowUpWideNarrow className="size-3.5" />
                        ) : null}
                    </button>
                ) : (
                    label
                )}
                <Popover onOpenChange={setOpen} open={open}>
                    <PopoverTrigger
                        render={
                            <button
                                aria-label={translate("Column options")}
                                className={cn("ml-auto shrink-0 rounded p-0.5", isFiltered && "text-primary")}
                                type="button"
                            >
                                <MoreVertical className="size-3.5" />
                            </button>
                        }
                    />
                    <PopoverContent align="end" className="flex w-64 flex-col gap-2">
                        {canSort ? (
                            <>
                                <div className="flex flex-col gap-1">
                                    <Button
                                        className="w-full"
                                        onClick={() => {
                                            column.toggleSorting(false);
                                            setOpen(false);
                                        }}
                                        size="sm"
                                        variant="outline"
                                    >
                                        <span className="grid w-full grid-cols-[1fr_2fr] items-center">
                                            <ArrowDownNarrowWide className="justify-self-center" />
                                            <span className="justify-self-center">{translate("Sort Ascending")}</span>
                                        </span>
                                    </Button>
                                    <Button
                                        className="w-full"
                                        onClick={() => {
                                            column.toggleSorting(true);
                                            setOpen(false);
                                        }}
                                        size="sm"
                                        variant="outline"
                                    >
                                        <span className="grid w-full grid-cols-[1fr_2fr] items-center">
                                            <ArrowUpWideNarrow className="justify-self-center" />
                                            <span className="justify-self-center">{translate("Sort Descending")}</span>
                                        </span>
                                    </Button>
                                </div>
                                {sortState ? (
                                    <Button
                                        className="w-full"
                                        onClick={() => {
                                            column.clearSorting();
                                            setOpen(false);
                                        }}
                                        size="sm"
                                        variant="outline"
                                    >
                                        {translate("Clear Sort")}
                                    </Button>
                                ) : null}
                                <DropdownMenuSeparator />
                            </>
                        ) : null}
                        {canFilter ? (
                            <>
                                <ColumnFilterPanel column={column} kind={kind} />
                                <DropdownMenuSeparator />
                            </>
                        ) : null}
                        <Button
                            onClick={() => {
                                onOpenManageColumns();
                                setOpen(false);
                            }}
                            size="sm"
                            variant="outline"
                        >
                            {translate("Manage Columns")}
                        </Button>
                        {canHide ? (
                            <Button
                                onClick={() => {
                                    column.toggleVisibility(false);
                                    setOpen(false);
                                }}
                                size="sm"
                                variant="outline"
                            >
                                {translate("Hide Column")}
                            </Button>
                        ) : null}
                    </PopoverContent>
                </Popover>
            </div>
            {column.getCanResize() ? (
                <div
                    className="absolute top-0 right-0 h-full w-1 cursor-col-resize touch-none select-none hover:bg-primary/50 data-[resizing]:bg-primary"
                    data-resizing={column.getIsResizing() || undefined}
                    onMouseDown={header.getResizeHandler()}
                    onTouchStart={header.getResizeHandler()}
                />
            ) : null}
        </>
    );
}

export { ColumnHeaderCell };
