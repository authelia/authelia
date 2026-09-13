import { fireEvent, render, screen } from "@testing-library/react";

import { Pagination } from "@components/UI/Pagination";

const onPageChange = vi.fn();
const onPageSizeChange = vi.fn();

beforeEach(() => {
    vi.clearAllMocks();
});

it("renders with the pagination data-slot and merges a passed className", () => {
    const { container } = render(
        <Pagination
            className="custom-pagination"
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            page={1}
            pageSize={25}
            total={120}
        />,
    );
    const el = container.querySelector('[data-slot="pagination"]');
    expect(el).toBeInTheDocument();
    expect(el).toHaveClass("custom-pagination");
});

it("renders the current range text", () => {
    render(
        <Pagination
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            page={1}
            pageSize={25}
            total={120}
        />,
    );
    expect(screen.getByText("1–25 of 120")).toBeInTheDocument();
});

it("renders the range for a middle page", () => {
    render(
        <Pagination
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            page={3}
            pageSize={25}
            total={120}
        />,
    );
    expect(screen.getByText("51–75 of 120")).toBeInTheDocument();
});

it("disables the previous button on the first page", () => {
    render(
        <Pagination
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            page={1}
            pageSize={25}
            total={120}
        />,
    );
    expect(document.getElementById("pagination-prev")).toBeDisabled();
    expect(document.getElementById("pagination-next")).not.toBeDisabled();
});

it("disables the next button on the last page", () => {
    render(
        <Pagination
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            page={5}
            pageSize={25}
            total={120}
        />,
    );
    expect(document.getElementById("pagination-next")).toBeDisabled();
    expect(document.getElementById("pagination-prev")).not.toBeDisabled();
});

it("calls onPageChange when clicking prev and next", () => {
    render(
        <Pagination
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            page={2}
            pageSize={25}
            total={120}
        />,
    );
    fireEvent.click(document.getElementById("pagination-prev")!);
    expect(onPageChange).toHaveBeenCalledWith(1);

    fireEvent.click(document.getElementById("pagination-next")!);
    expect(onPageChange).toHaveBeenCalledWith(3);
});

it("calls onPageSizeChange when a new page size is selected", () => {
    render(
        <Pagination
            onPageChange={onPageChange}
            onPageSizeChange={onPageSizeChange}
            page={1}
            pageSize={25}
            total={120}
        />,
    );
    fireEvent.click(document.getElementById("pagination-page-size")!);
    fireEvent.click(screen.getByText("50 / page"));
    expect(onPageSizeChange).toHaveBeenCalledWith(50);
});
