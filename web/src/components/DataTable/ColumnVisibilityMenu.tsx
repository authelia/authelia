import { SlidersHorizontal } from "lucide-react";

import type { ColumnDef } from "@components/DataTable/DataTable";
import { Button } from "@components/UI/Button";
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuTrigger,
} from "@components/UI/DropdownMenu";

function storageKey(id: string): string {
    return `datatable:${id}:columns`;
}

function loadColumnVisibility<T>(id: string, columns: ColumnDef<T>[]): Record<string, boolean> {
    const defaults: Record<string, boolean> = {};

    for (const column of columns) {
        defaults[column.field] = !column.hidden;
    }

    try {
        const raw = window.localStorage.getItem(storageKey(id));

        if (!raw) {
            return defaults;
        }

        const stored = JSON.parse(raw) as Record<string, boolean>;

        return { ...defaults, ...stored };
    } catch {
        return defaults;
    }
}

function persistColumnVisibility(id: string, visibility: Record<string, boolean>): void {
    try {
        window.localStorage.setItem(storageKey(id), JSON.stringify(visibility));
    } catch {
        // localStorage can throw or be unavailable (private browsing, quota exceeded, disabled); in
        // that case visibility simply does not persist across reloads.
    }
}

interface ColumnVisibilityMenuProps<T> {
    id: string;
    columns: ColumnDef<T>[];
    visibility: Record<string, boolean>;
    onChange: (visibility: Record<string, boolean>) => void;
}

function ColumnVisibilityMenu<T>({ columns, id, onChange, visibility }: Readonly<ColumnVisibilityMenuProps<T>>) {
    const handleToggle = (field: string, checked: boolean) => {
        const next = { ...visibility, [field]: checked };

        onChange(next);
        persistColumnVisibility(id, next);
    };

    return (
        <DropdownMenu>
            <DropdownMenuTrigger
                render={
                    <Button aria-label="Toggle columns" id={`${id}-columns`} size="icon" variant="outline">
                        <SlidersHorizontal />
                    </Button>
                }
            />
            <DropdownMenuContent align="end">
                {columns.map((column) => (
                    <DropdownMenuCheckboxItem
                        checked={visibility[column.field] ?? true}
                        closeOnClick={false}
                        key={column.field}
                        onCheckedChange={(checked) => handleToggle(column.field, checked)}
                    >
                        {column.header}
                    </DropdownMenuCheckboxItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

export { ColumnVisibilityMenu, loadColumnVisibility, persistColumnVisibility, storageKey };
export type { ColumnVisibilityMenuProps };
