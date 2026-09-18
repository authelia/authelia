import {
    columnFacetingFeature,
    columnFilteringFeature,
    columnOrderingFeature,
    columnResizingFeature,
    columnSizingFeature,
    columnVisibilityFeature,
    createFacetedMinMaxValues,
    createFacetedRowModel,
    createFacetedUniqueValues,
    createFilteredRowModel,
    createPaginatedRowModel,
    createSortedRowModel,
    globalFilteringFeature,
    rowPaginationFeature,
    rowSortingFeature,
    sortFn_alphanumeric,
    sortFn_basic,
    tableFeatures,
} from "@tanstack/react-table";

import { columnFilterFn, globalFilterFn } from "@components/DataTable/filters";

type ColumnKind = "date" | "number" | "string";

interface DataTableColumnMeta {
    display: (row: any) => string;
    kind: ColumnKind;
}

const dataTableFeatures = tableFeatures({
    columnFacetingFeature,
    columnFilteringFeature,
    columnMeta: {} as DataTableColumnMeta,
    columnOrderingFeature,
    columnResizingFeature,
    columnSizingFeature,
    columnVisibilityFeature,
    facetedMinMaxValues: createFacetedMinMaxValues(),
    facetedRowModel: createFacetedRowModel(),
    facetedUniqueValues: createFacetedUniqueValues(),
    filteredRowModel: createFilteredRowModel(),
    filterFns: { columnFilter: columnFilterFn, globalSearch: globalFilterFn },
    globalFilteringFeature,
    paginatedRowModel: createPaginatedRowModel(),
    rowPaginationFeature,
    rowSortingFeature,
    sortedRowModel: createSortedRowModel(),
    sortFns: { alphanumeric: sortFn_alphanumeric, basic: sortFn_basic },
});

type DataTableFeatures = typeof dataTableFeatures;

export { dataTableFeatures };
export type { ColumnKind, DataTableColumnMeta, DataTableFeatures };
