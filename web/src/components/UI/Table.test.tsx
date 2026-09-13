import { render, screen } from "@testing-library/react";

import {
    Table,
    TableBody,
    TableCaption,
    TableCell,
    TableFooter,
    TableHead,
    TableHeader,
    TableRow,
} from "@components/UI/Table";

function renderTable() {
    return render(
        <Table className="custom-table">
            <TableCaption>Caption</TableCaption>
            <TableHeader>
                <TableRow>
                    <TableHead>Name</TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                <TableRow>
                    <TableCell>Alice</TableCell>
                </TableRow>
            </TableBody>
            <TableFooter>
                <TableRow>
                    <TableCell>Total</TableCell>
                </TableRow>
            </TableFooter>
        </Table>,
    );
}

it("wraps the table in a data-slot table-container", () => {
    const { container } = renderTable();
    const wrapper = container.querySelector('[data-slot="table-container"]');
    expect(wrapper).toBeInTheDocument();
    expect(wrapper).toHaveClass("relative", "w-full", "overflow-x-auto");
    expect(wrapper?.querySelector("table")).toBe(container.querySelector('[data-slot="table"]'));
});

it("merges a passed className on the table element", () => {
    const { container } = renderTable();
    expect(container.querySelector('[data-slot="table"]')).toHaveClass("custom-table");
});

it("renders the header, body, footer and caption with their data-slot attributes", () => {
    const { container } = renderTable();
    expect(container.querySelector('[data-slot="table-header"]')).toBeInTheDocument();
    expect(container.querySelector('[data-slot="table-body"]')).toBeInTheDocument();
    expect(container.querySelector('[data-slot="table-footer"]')).toBeInTheDocument();
    expect(container.querySelector('[data-slot="table-caption"]')).toBeInTheDocument();
    expect(container.querySelectorAll('[data-slot="table-row"]')).toHaveLength(3);
    expect(container.querySelector('[data-slot="table-head"]')).toBeInTheDocument();
    expect(container.querySelectorAll('[data-slot="table-cell"]')).toHaveLength(2);
});

it("renders the provided content", () => {
    renderTable();
    expect(screen.getByText("Alice")).toBeInTheDocument();
    expect(screen.getByText("Total")).toBeInTheDocument();
    expect(screen.getByText("Caption")).toBeInTheDocument();
});
