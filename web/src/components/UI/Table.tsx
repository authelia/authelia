import { type ComponentProps } from "react";

import { cn } from "@utils/Styles";

function Table({ className, ...props }: ComponentProps<"table">) {
    return (
        <div data-slot="table-container" className="relative w-full overflow-x-auto">
            <table data-slot="table" className={cn("w-full caption-bottom text-sm", className)} {...props} />
        </div>
    );
}

function TableHeader({ className, ...props }: ComponentProps<"thead">) {
    return <thead data-slot="table-header" className={cn("[&_tr]:border-b", className)} {...props} />;
}

function TableBody({ className, ...props }: ComponentProps<"tbody">) {
    return <tbody data-slot="table-body" className={cn("[&_tr:last-child]:border-0", className)} {...props} />;
}

function TableFooter({ className, ...props }: ComponentProps<"tfoot">) {
    return (
        <tfoot
            data-slot="table-footer"
            className={cn("border-t bg-muted/50 font-medium [&>tr]:last:border-b-0", className)}
            {...props}
        />
    );
}

function TableRow({ className, ...props }: ComponentProps<"tr">) {
    return (
        <tr
            data-slot="table-row"
            className={cn("border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted", className)}
            {...props}
        />
    );
}

function TableHead({ className, ...props }: ComponentProps<"th">) {
    return (
        <th
            data-slot="table-head"
            className={cn(
                "h-10 px-2 text-left align-middle font-medium whitespace-nowrap text-foreground [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]",
                className,
            )}
            {...props}
        />
    );
}

function TableCell({ className, ...props }: ComponentProps<"td">) {
    return (
        <td
            data-slot="table-cell"
            className={cn(
                "p-2 align-middle whitespace-nowrap [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]",
                className,
            )}
            {...props}
        />
    );
}

function TableCaption({ className, ...props }: ComponentProps<"caption">) {
    return (
        <caption data-slot="table-caption" className={cn("mt-4 text-sm text-muted-foreground", className)} {...props} />
    );
}

export { Table, TableHeader, TableBody, TableFooter, TableRow, TableHead, TableCell, TableCaption };
