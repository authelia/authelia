import { useCallback, useMemo, useState } from "react";

import { type ColumnFiltersState, type SortingState } from "@tanstack/react-table";

import { persistentStorage } from "@hooks/PersistentStorage";

interface PersistableColumn {
    field: string;
    hidden?: boolean;
}

interface PersistedTableState {
    columnFilters: ColumnFiltersState;
    columnOrder: string[];
    columnSizing: Record<string, number>;
    columnVisibility: Record<string, boolean>;
    globalFilter: string;
    pageSize: number;
    sorting: SortingState;
}

function storageKey(id: string): string {
    return `datatable:${id}:v2`;
}

function defaultState(
    columns: PersistableColumn[],
    initialSort: SortingState,
    initialPageSize: number,
): PersistedTableState {
    const columnVisibility: Record<string, boolean> = {};

    for (const column of columns) {
        columnVisibility[column.field] = !column.hidden;
    }

    return {
        columnFilters: [],
        columnOrder: columns.map((column) => column.field),
        columnSizing: {},
        columnVisibility,
        globalFilter: "",
        pageSize: initialPageSize,
        sorting: initialSort,
    };
}

function pruneToKnownColumns(state: PersistedTableState, columns: PersistableColumn[]): PersistedTableState {
    const fields = new Set(columns.map((column) => column.field));
    const knownOrder = state.columnOrder.filter((field) => fields.has(field));
    const missingFromOrder = columns.map((column) => column.field).filter((field) => !knownOrder.includes(field));

    return {
        ...state,
        columnFilters: state.columnFilters.filter((filter) => fields.has(filter.id)),
        columnOrder: [...knownOrder, ...missingFromOrder],
        columnSizing: Object.fromEntries(Object.entries(state.columnSizing).filter(([field]) => fields.has(field))),
        columnVisibility: Object.fromEntries(
            Object.entries(state.columnVisibility).filter(([field]) => fields.has(field)),
        ),
    };
}

function loadState(
    id: string,
    columns: PersistableColumn[],
    initialSort: SortingState,
    initialPageSize: number,
): PersistedTableState {
    const defaults = defaultState(columns, initialSort, initialPageSize);
    const stored = persistentStorage.getItem(storageKey(id)) as Partial<PersistedTableState> | undefined;

    if (!stored || typeof stored !== "object") {
        return defaults;
    }

    return pruneToKnownColumns({ ...defaults, ...stored }, columns);
}

function persistState(id: string, state: PersistedTableState): void {
    persistentStorage.setItem(storageKey(id), state);
}

function useTableState(id: string, columns: PersistableColumn[], initialSort: SortingState, initialPageSize: number) {
    const [state, setState] = useState<PersistedTableState>(() => loadState(id, columns, initialSort, initialPageSize));

    const update = useCallback(
        (updater: (previous: PersistedTableState) => PersistedTableState) => {
            setState((previous) => {
                const next = updater(previous);

                persistState(id, next);

                return next;
            });
        },
        [id],
    );

    const reset = useCallback(() => {
        update((previous) => ({ ...previous, columnFilters: [], globalFilter: "" }));
    }, [update]);

    return useMemo(() => ({ reset, state, update }), [reset, state, update]);
}

export { storageKey, useTableState };
export type { PersistedTableState };
