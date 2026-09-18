import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { Pencil, Trash2 } from "lucide-react";

import { type ColumnDef, DataTable, type RowAction } from "@components/DataTable/DataTable";

interface Row {
    username: string;
    displayName: string;
}

const rows: Row[] = [
    { displayName: "Bob B", username: "bob" },
    { displayName: "Alice A", username: "alice" },
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
        getRowId: (row: Row) => row.username,
        id: "user-management-table",
        rowClassPrefix: "user-row-",
        rows,
    };
}

function getColumnOptionsButton(headerText: string): HTMLElement {
    const headerCell = screen.getByText(headerText).closest("th")!;

    return within(headerCell).getByLabelText("Column options");
}

function tableBodyRowTexts(): string[] {
    const table = document.getElementById("user-management-table")!;

    return within(table)
        .getAllByRole("row")
        .filter((row) => row.querySelector("td"))
        .map((row) => row.textContent ?? "");
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

it("shows a hidden column once made visible via the Manage Columns dialog", async () => {
    render(<DataTable {...baseProps()} />);

    fireEvent.click(getColumnOptionsButton("Username"));
    fireEvent.click(await screen.findByText("Manage Columns"));
    fireEvent.click(await screen.findByRole("checkbox", { name: "Secret" }));

    expect(screen.getAllByText("hidden-value")).toHaveLength(rows.length);
});

it("persists column visibility changes to localStorage under the v2 key", async () => {
    render(<DataTable {...baseProps()} />);

    fireEvent.click(getColumnOptionsButton("Username"));
    fireEvent.click(await screen.findByText("Hide Column"));

    const stored = JSON.parse(window.localStorage.getItem("datatable:user-management-table:v2")!);

    expect(stored.columnVisibility.username).toBe(false);
});

it("clicking a sortable header cycles sort asc -> desc -> none", () => {
    render(<DataTable {...baseProps()} />);

    expect(tableBodyRowTexts()[0]).toContain("bob");

    fireEvent.click(screen.getByText("Username"));
    expect(tableBodyRowTexts()[0]).toContain("alice");

    fireEvent.click(screen.getByText("Username"));
    expect(tableBodyRowTexts()[0]).toContain("bob");

    fireEvent.click(screen.getByText("Username"));
    expect(tableBodyRowTexts()[0]).toContain("bob");
});

it("clicking a different column's header sorts that column ascending", () => {
    render(<DataTable {...baseProps()} />);

    fireEvent.click(screen.getByText("Display Name"));

    expect(tableBodyRowTexts()[0]).toContain("Alice A");
});

it("debounces the search input, narrowing rows only after 250ms", () => {
    vi.useFakeTimers();
    render(<DataTable {...baseProps()} />);

    const input = document.getElementById("user-management-table-search")!;

    fireEvent.change(input, { target: { value: "alice" } });
    act(() => vi.advanceTimersByTime(249));
    expect(tableBodyRowTexts()).toHaveLength(2);

    act(() => vi.advanceTimersByTime(1));
    expect(tableBodyRowTexts()).toHaveLength(1);
    expect(tableBodyRowTexts()[0]).toContain("alice");

    vi.useRealTimers();
});

it("only the latest search value takes effect after rapid typing", () => {
    vi.useFakeTimers();
    render(<DataTable {...baseProps()} />);

    const input = document.getElementById("user-management-table-search")!;

    fireEvent.change(input, { target: { value: "b" } });
    act(() => vi.advanceTimersByTime(100));
    fireEvent.change(input, { target: { value: "bob" } });
    act(() => vi.advanceTimersByTime(250));

    expect(tableBodyRowTexts()).toHaveLength(1);
    expect(tableBodyRowTexts()[0]).toContain("bob");

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

    expect(rowActions[0].onClick).toHaveBeenCalledWith(rows[1]);
});

it("double-clicking a row calls onRowDoubleClick with that row's data", () => {
    const onRowDoubleClick = vi.fn();

    render(<DataTable {...baseProps()} onRowDoubleClick={onRowDoubleClick} />);

    fireEvent.doubleClick(screen.getByText("alice"));

    expect(onRowDoubleClick).toHaveBeenCalledWith(rows[1]);
});

it("shows the empty text when there are no rows and no filters are active", () => {
    render(<DataTable {...baseProps()} rows={[]} />);

    expect(screen.getByText("No rows")).toBeInTheDocument();
});

it("shows a loading spinner overlay when loading", () => {
    const { container } = render(<DataTable {...baseProps()} loading />);

    expect(container.querySelector('[data-slot="spinner"]')).toBeInTheDocument();
});

it("shows a filtered-empty hint with a Clear Filters action when a filter matches nothing", () => {
    vi.useFakeTimers();
    render(<DataTable {...baseProps()} />);

    const input = document.getElementById("user-management-table-search")!;

    fireEvent.change(input, { target: { value: "nobody-matches-this" } });
    act(() => vi.advanceTimersByTime(250));

    expect(screen.getByText("There are no results for the current filter(s).")).toBeInTheDocument();

    fireEvent.click(screen.getByText("Clear Filters"));

    expect(tableBodyRowTexts()).toHaveLength(2);

    vi.useRealTimers();
});

it("narrows rows using a per-column values filter", async () => {
    render(<DataTable {...baseProps()} />);

    fireEvent.click(getColumnOptionsButton("Username"));

    const popover = (await screen.findByPlaceholderText("Search values")).closest(
        '[data-slot="popover-content"]',
    ) as HTMLElement;
    const checkbox = within(popover).getByRole("checkbox", { name: "bob" });

    fireEvent.click(checkbox);

    expect(tableBodyRowTexts()).toHaveLength(1);
    expect(tableBodyRowTexts()[0]).toContain("bob");
});

it("narrows the values checklist itself using the mini search field, without filtering rows", async () => {
    render(<DataTable {...baseProps()} />);

    fireEvent.click(getColumnOptionsButton("Username"));

    const search = await screen.findByPlaceholderText("Search values");
    const popover = search.closest('[data-slot="popover-content"]') as HTMLElement;

    expect(within(popover).getByText("alice")).toBeInTheDocument();
    expect(within(popover).getByText("bob")).toBeInTheDocument();

    fireEvent.change(search, { target: { value: "ali" } });

    expect(within(popover).getByText("alice")).toBeInTheDocument();
    expect(within(popover).queryByText("bob")).not.toBeInTheDocument();
    expect(tableBodyRowTexts()).toHaveLength(2);
});

it("moves a column via the Manage Columns dialog", async () => {
    render(<DataTable {...baseProps()} />);

    fireEvent.click(getColumnOptionsButton("Username"));
    fireEvent.click(await screen.findByText("Manage Columns"));
    fireEvent.click((await screen.findAllByRole("button", { name: "Move Down" }))[0]);

    const headers = within(document.getElementById("user-management-table")!)
        .getAllByRole("columnheader", { hidden: true })
        .map((header) => header.textContent);

    expect(headers[0]).toContain("Display Name");
    expect(headers[1]).toContain("Username");
});

it("only renders a bounded window of rows out of 500", () => {
    const manyRows: Row[] = Array.from({ length: 500 }, (_, index) => ({
        displayName: `Display ${index}`,
        username: `user${index}`,
    }));

    render(<DataTable {...baseProps()} initialPageSize={500} rows={manyRows} />);

    const table = document.getElementById("user-management-table")!;
    const renderedDataRows = within(table)
        .getAllByRole("row")
        .filter((row) => row.querySelector("td")?.textContent?.startsWith("user"));

    expect(renderedDataRows.length).toBeLessThan(100);
    expect(renderedDataRows.length).toBeGreaterThan(0);
});
