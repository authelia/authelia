import { useState } from "react";

import { type Column, type RowData } from "@tanstack/react-table";
import { useTranslation } from "react-i18next";

import { type ColumnKind, type DataTableFeatures } from "@components/DataTable/features";
import { type ColumnFilterValue, isFilterValueEmpty } from "@components/DataTable/filters";
import { Checkbox } from "@components/UI/Checkbox";
import { Input } from "@components/UI/Input";

const MAX_VALUE_LENGTH = 40;
const EMPTY_VALUE_PLACEHOLDER = "-";
const OPTIONS_PAGE_SIZE = 200;

interface ColumnFilterPanelProps<T extends RowData> {
    column: Column<DataTableFeatures, T, unknown>;
    kind: ColumnKind;
}

function truncate(value: string): string {
    return value.length > MAX_VALUE_LENGTH ? `${value.slice(0, MAX_VALUE_LENGTH)}…` : value;
}

function toDateInputValue(ms: null | number): string {
    return ms == null ? "" : new Date(ms).toISOString().slice(0, 10);
}

function fromDateInputValue(input: string): null | number {
    return input === "" ? null : new Date(input).getTime();
}

function ColumnFilterPanel<T extends RowData>({ column, kind }: Readonly<ColumnFilterPanelProps<T>>) {
    const { t: translate } = useTranslation("settings");
    const filterValue = column.getFilterValue() as ColumnFilterValue | undefined;

    const [search, setSearch] = useState("");
    const [loadedCount, setLoadedCount] = useState(OPTIONS_PAGE_SIZE);

    if (kind === "date") {
        const value =
            filterValue?.mode === "dateRange" ? filterValue : { from: null, mode: "dateRange" as const, to: null };

        return (
            <div className="flex flex-col gap-2">
                <label className="flex flex-col gap-1 text-xs text-muted-foreground">
                    {translate("From")}
                    <Input
                        className="h-9"
                        onChange={(event) =>
                            column.setFilterValue({
                                ...value,
                                from: fromDateInputValue(event.target.value),
                            } satisfies ColumnFilterValue)
                        }
                        type="date"
                        value={toDateInputValue(value.from)}
                    />
                </label>
                <label className="flex flex-col gap-1 text-xs text-muted-foreground">
                    {translate("To")}
                    <Input
                        className="h-9"
                        onChange={(event) =>
                            column.setFilterValue({
                                ...value,
                                to: fromDateInputValue(event.target.value),
                            } satisfies ColumnFilterValue)
                        }
                        type="date"
                        value={toDateInputValue(value.to)}
                    />
                </label>
                {!isFilterValueEmpty(value) ? (
                    <button
                        className="self-start text-xs text-muted-foreground underline"
                        onClick={() => column.setFilterValue(undefined)}
                        type="button"
                    >
                        {translate("Clear Filter")}
                    </button>
                ) : null}
            </div>
        );
    }

    const selectedValues = filterValue?.mode === "values" ? filterValue.values : [];
    const facets = column.getFacetedUniqueValues() as Map<string, number>;
    const allOptions = Array.from(facets.keys())
        .filter((option) => option !== EMPTY_VALUE_PLACEHOLDER)
        .sort((a, b) => a.localeCompare(b));
    const loadedOptions = allOptions.slice(0, loadedCount);
    const options = search.trim()
        ? loadedOptions.filter((option) => option.toLowerCase().includes(search.trim().toLowerCase()))
        : loadedOptions;
    const hasMore = loadedOptions.length < allOptions.length;

    const toggleValue = (option: string, checked: boolean) => {
        const next = checked ? [...selectedValues, option] : selectedValues.filter((value) => value !== option);

        column.setFilterValue({ mode: "values", values: next } satisfies ColumnFilterValue);
    };

    return (
        <div className="flex flex-col gap-2">
            <Input
                className="h-9"
                onChange={(event) => setSearch(event.target.value)}
                placeholder={translate("Search values")}
                value={search}
            />
            <div className="flex max-h-48 flex-col gap-1 overflow-y-auto">
                {options.length === 0 ? (
                    <span className="px-1 text-center text-xs text-muted-foreground">
                        {translate(search.trim() ? "No matches" : "No values")}
                    </span>
                ) : (
                    options.map((option, index) => {
                        const checkboxId = `${column.id}-filter-value-${index}`;

                        return (
                            <div className="flex items-center gap-2 px-1 py-0.5 text-sm" key={option} title={option}>
                                <Checkbox
                                    checked={selectedValues.includes(option)}
                                    id={checkboxId}
                                    onCheckedChange={(checked) => toggleValue(option, checked === true)}
                                />
                                <label className="flex-1 truncate" htmlFor={checkboxId}>
                                    {truncate(option)}
                                </label>
                                <span className="text-xs text-muted-foreground">{facets.get(option)}</span>
                            </div>
                        );
                    })
                )}
                {hasMore ? (
                    <button
                        className="self-center px-1 py-0.5 text-xs text-muted-foreground underline"
                        onClick={() => setLoadedCount((previous) => previous + OPTIONS_PAGE_SIZE)}
                        type="button"
                    >
                        {translate("Load More")}
                    </button>
                ) : null}
            </div>
            {filterValue && !isFilterValueEmpty(filterValue) ? (
                <button
                    className="self-start text-xs text-muted-foreground underline"
                    onClick={() => column.setFilterValue(undefined)}
                    type="button"
                >
                    {translate("Clear Filter")}
                </button>
            ) : null}
        </div>
    );
}

export { ColumnFilterPanel };
