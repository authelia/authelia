import { type FilterFn } from "@tanstack/react-table";

interface DateRangeFilterValue {
    from: null | number;
    mode: "dateRange";
    to: null | number;
}

interface ValuesFilterValue {
    mode: "values";
    values: string[];
}

type ColumnFilterValue = DateRangeFilterValue | ValuesFilterValue;

function isFilterValueEmpty(value: unknown): boolean {
    const filter = value as ColumnFilterValue | undefined;

    if (!filter) {
        return true;
    }

    switch (filter.mode) {
        case "dateRange":
            return filter.from == null && filter.to == null;
        case "values":
            return filter.values.length === 0;
    }
}

const columnFilterFn: FilterFn<any, any> = (row, columnId, filterValue) => {
    const filter = filterValue as ColumnFilterValue;

    switch (filter.mode) {
        case "dateRange": {
            const raw = row.getValue<null | number>(columnId);

            if (raw == null) {
                return false;
            }

            if (filter.from != null && raw < filter.from) {
                return false;
            }

            return !(filter.to != null && raw > filter.to);
        }
        case "values": {
            const display = String(row.getValue(columnId) ?? "");

            return filter.values.includes(display);
        }
    }
};

columnFilterFn.autoRemove = isFilterValueEmpty;

const globalFilterFn: FilterFn<any, any> = (row, columnId, filterValue) => {
    const search = String(filterValue ?? "")
        .trim()
        .toLowerCase();

    if (!search) {
        return true;
    }

    const column = row.table.getColumn(columnId);
    const meta = column?.columnDef.meta as { display?: (row: unknown) => string } | undefined;
    const display = meta?.display?.(row.original) ?? String(row.getValue(columnId) ?? "");

    return display.toLowerCase().includes(search);
};

export { columnFilterFn, globalFilterFn, isFilterValueEmpty };
export type { ColumnFilterValue, DateRangeFilterValue, ValuesFilterValue };
