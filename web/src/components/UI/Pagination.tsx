import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react";

import { Button } from "@components/UI/Button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@components/UI/DropdownMenu";
import { cn } from "@utils/Styles";

interface PaginationProps {
    page: number;
    pageSize: number;
    total: number;
    pageSizeOptions?: number[];
    onPageChange: (page: number) => void;
    onPageSizeChange: (size: number) => void;
    className?: string;
}

function Pagination({
    className,
    onPageChange,
    onPageSizeChange,
    page,
    pageSize,
    pageSizeOptions = [10, 25, 50, 100],
    total,
}: PaginationProps) {
    const lastPage = Math.max(1, Math.ceil(total / pageSize));
    const rangeStart = total === 0 ? 0 : (page - 1) * pageSize + 1;
    const rangeEnd = Math.min(total, page * pageSize);

    return (
        <div
            data-slot="pagination"
            className={cn("flex flex-wrap items-center justify-between gap-4 text-sm", className)}
        >
            <div data-slot="pagination-range" className="text-muted-foreground">
                {rangeStart}–{rangeEnd} of {total}
            </div>
            <div className="flex items-center gap-4">
                <DropdownMenu>
                    <DropdownMenuTrigger
                        render={
                            <Button id="pagination-page-size" size="sm" variant="outline">
                                {pageSize} / page
                            </Button>
                        }
                    />
                    <DropdownMenuContent align="end">
                        {pageSizeOptions.map((option) => (
                            <DropdownMenuItem key={option} onClick={() => onPageSizeChange(option)}>
                                {option} / page
                            </DropdownMenuItem>
                        ))}
                    </DropdownMenuContent>
                </DropdownMenu>
                <div className="flex items-center gap-1">
                    <Button
                        aria-label="Previous page"
                        disabled={page <= 1}
                        id="pagination-prev"
                        onClick={() => onPageChange(page - 1)}
                        size="icon"
                        variant="outline"
                    >
                        <ChevronLeftIcon />
                    </Button>
                    <Button
                        aria-label="Next page"
                        disabled={page >= lastPage}
                        id="pagination-next"
                        onClick={() => onPageChange(page + 1)}
                        size="icon"
                        variant="outline"
                    >
                        <ChevronRightIcon />
                    </Button>
                </div>
            </div>
        </div>
    );
}

export { Pagination };
