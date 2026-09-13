import { fireEvent, render, screen, within } from "@testing-library/react";
import { Pencil, Trash2 } from "lucide-react";

import { type ColumnDef, DataTable, type RowAction } from "@components/DataTable/DataTable";

interface Row {
    username: string;
    displayName: string;
}

const rows: Row[] = [
    { displayName: "Alice A", username: "alice" },
    { displayName: "Bob B", username: "bob" },
];

const columns: ColumnDef<Row>[] = [
    { field: "username", header: "Username", value: (row) => row.username },
    { field: "displayName", header: "Display Name", value: (row) => row.displayName },
    { field: "secret", header: "Secret", hidden: true, value: () => "hidden-value" },
];

const rowActions: RowAction<Row>[] = [
    { icon: <Pencil data-testid="edit-icon" />, id: () => "edit", label: "Edit", onClick: vi.fn() },
    { icon: <Trash2 data-testid="delete-icon" />, id: () => "delete", label: "Delete", onClick: vi.fn() },
];

function baseProps() {
    return {
        columns,
        emptyText: "No rows",
        filter: "",
        getRowId: (row: Row) => row.username,
        id: "user-management-table",
        onFilterChange: vi.fn(),
        onPaginationChange: vi.fn(),
        onSortChange: vi.fn(),
        pagination: { page: 1, pageSize: 25, total: rows.length },
        rowClassPrefix: "user-row-",
        rows,
        sort: { direction: "asc" as const, field: "username" },
    };
}

beforeEach(() => {
    window.localStorage.clear();
    vi.clearAllMocks();
});

it("renders one header cell per visible column and hides an initially-hidden column", () => {
    render(<DataTable {...baseProps()} />);

    expect(screen.getByText("Username")).toBeInTheDocument();
    expect(screen.getByText("Display Name")).toBeInTheDocument();
    expect(screen.queryByText("Secret")).not.toBeInTheDocument();
});

it("shows a hidden column once toggled visible via the column menu", async () => {
    render(<DataTable {...baseProps()} />);

    fireEvent.click(document.getElementById("user-management-table-columns")!);
    fireEvent.click(await screen.findByText("Secret"));

    expect(screen.getAllByText("hidden-value")).toHaveLength(rows.length);
});

it("persists column visibility toggles to localStorage", async () => {
    render(<DataTable {...baseProps()} />);

    fireEvent.click(document.getElementById("user-management-table-columns")!);
    fireEvent.click(await screen.findByRole("menuitemcheckbox", { name: "Username" }));

    const stored = JSON.parse(window.localStorage.getItem("datatable:user-management-table:columns")!);

    expect(stored.username).toBe(false);
});

it("clicking a sortable header toggles the sort direction, asc -> desc -> asc", () => {
    const onSortChange = vi.fn();
    const { rerender } = render(<DataTable {...baseProps()} onSortChange={onSortChange} />);

    fireEvent.click(screen.getByText("Username"));
    expect(onSortChange).toHaveBeenLastCalledWith({ direction: "desc", field: "username" });

    rerender(
        <DataTable {...baseProps()} onSortChange={onSortChange} sort={{ direction: "desc", field: "username" }} />,
    );
    fireEvent.click(screen.getByText("Username"));
    expect(onSortChange).toHaveBeenLastCalledWith({ direction: "asc", field: "username" });
});

it("clicking a different column's header sorts that column ascending", () => {
    const onSortChange = vi.fn();
    render(<DataTable {...baseProps()} onSortChange={onSortChange} />);

    fireEvent.click(screen.getByText("Display Name"));
    expect(onSortChange).toHaveBeenCalledWith({ direction: "asc", field: "displayName" });
});

it("debounces the search input, firing onFilterChange only after 250ms", () => {
    vi.useFakeTimers();
    const onFilterChange = vi.fn();
    render(<DataTable {...baseProps()} onFilterChange={onFilterChange} />);

    const input = document.getElementById("user-management-table-search")!;
    fireEvent.change(input, { target: { value: "al" } });

    vi.advanceTimersByTime(249);
    expect(onFilterChange).not.toHaveBeenCalled();

    vi.advanceTimersByTime(1);
    expect(onFilterChange).toHaveBeenCalledWith("al");
    expect(onFilterChange).toHaveBeenCalledTimes(1);

    vi.useRealTimers();
});

it("only the latest search value fires after rapid typing", () => {
    vi.useFakeTimers();
    const onFilterChange = vi.fn();
    render(<DataTable {...baseProps()} onFilterChange={onFilterChange} />);

    const input = document.getElementById("user-management-table-search")!;
    fireEvent.change(input, { target: { value: "a" } });
    vi.advanceTimersByTime(100);
    fireEvent.change(input, { target: { value: "al" } });
    vi.advanceTimersByTime(250);

    expect(onFilterChange).toHaveBeenCalledTimes(1);
    expect(onFilterChange).toHaveBeenCalledWith("al");

    vi.useRealTimers();
});

it("renders row action buttons with ids derived from the row-class-prefix, row id and action id", () => {
    render(<DataTable {...baseProps()} rowActions={rowActions} />);

    expect(document.getElementById("user-row-alice-edit")).toBeInTheDocument();
    expect(document.getElementById("user-row-alice-delete")).toBeInTheDocument();
    expect(document.getElementById("user-row-bob-edit")).toBeInTheDocument();
    expect(document.getElementById("user-row-bob-delete")).toBeInTheDocument();
});

it("clicking a row action invokes its onClick with the row", () => {
    render(<DataTable {...baseProps()} rowActions={rowActions} />);

    fireEvent.click(document.getElementById("user-row-alice-edit")!);

    expect(rowActions[0].onClick).toHaveBeenCalledWith(rows[0]);
});

it("double-clicking a row calls onRowDoubleClick with that row's data", () => {
    const onRowDoubleClick = vi.fn();
    render(<DataTable {...baseProps()} onRowDoubleClick={onRowDoubleClick} />);

    fireEvent.doubleClick(screen.getByText("alice"));

    expect(onRowDoubleClick).toHaveBeenCalledWith(rows[0]);
});

it("shows the empty text when there are no rows", () => {
    render(<DataTable {...baseProps()} rows={[]} />);

    expect(screen.getByText("No rows")).toBeInTheDocument();
});

it("shows a loading spinner overlay when loading", () => {
    const { container } = render(<DataTable {...baseProps()} loading />);

    expect(container.querySelector('[data-slot="spinner"]')).toBeInTheDocument();
});

it("only renders a bounded window of rows out of 500", () => {
    const manyRows: Row[] = Array.from({ length: 500 }, (_, index) => ({
        displayName: `Display ${index}`,
        username: `user${index}`,
    }));

    render(
        <DataTable
            {...baseProps()}
            getRowId={(row) => row.username}
            pagination={{ page: 1, pageSize: 500, total: 500 }}
            rows={manyRows}
        />,
    );

    const table = document.getElementById("user-management-table")!;
    const renderedDataRows = within(table)
        .getAllByRole("row")
        .filter((row) => row.querySelector("td")?.textContent?.startsWith("user"));

    expect(renderedDataRows.length).toBeLessThan(100);
    expect(renderedDataRows.length).toBeGreaterThan(0);
});
