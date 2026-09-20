import { type ComponentProps } from "react";

import { type VariantProps, cva } from "class-variance-authority";

import { cn } from "@utils/Styles";

function Empty({ className, ...props }: ComponentProps<"div">) {
    return (
        <div
            data-slot="empty"
            className={cn(
                "flex min-w-0 flex-1 flex-col items-center justify-center gap-6 rounded-lg border-dashed p-6 text-center text-balance md:p-12",
                className,
            )}
            {...props}
        />
    );
}

function EmptyHeader({ className, ...props }: ComponentProps<"div">) {
    return (
        <div
            data-slot="empty-header"
            className={cn("flex max-w-sm flex-col items-center gap-2 text-center", className)}
            {...props}
        />
    );
}

const emptyMediaVariants = cva(
    "mb-2 flex shrink-0 items-center justify-center [&_svg]:pointer-events-none [&_svg]:shrink-0",
    {
        defaultVariants: {
            variant: "default",
        },
        variants: {
            variant: {
                default: "bg-transparent",
                icon: "flex size-10 shrink-0 items-center justify-center rounded-lg bg-muted text-foreground [&_svg:not([class*='size-'])]:size-6",
            },
        },
    },
);

function EmptyMedia({
    className,
    variant = "default",
    ...props
}: ComponentProps<"div"> & VariantProps<typeof emptyMediaVariants>) {
    return (
        <div
            data-slot="empty-icon"
            data-variant={variant}
            className={cn(emptyMediaVariants({ className, variant }))}
            {...props}
        />
    );
}

function EmptyTitle({ className, ...props }: ComponentProps<"div">) {
    return <div data-slot="empty-title" className={cn("text-lg font-medium tracking-tight", className)} {...props} />;
}

function EmptyDescription({ className, ...props }: ComponentProps<"p">) {
    return (
        <div
            data-slot="empty-description"
            className={cn(
                "text-sm/relaxed text-muted-foreground [&>a]:underline [&>a]:underline-offset-4 [&>a:hover]:text-primary",
                className,
            )}
            {...props}
        />
    );
}

function EmptyContent({ className, ...props }: ComponentProps<"div">) {
    return (
        <div
            data-slot="empty-content"
            className={cn("flex w-full max-w-sm min-w-0 flex-col items-center gap-4 text-sm text-balance", className)}
            {...props}
        />
    );
}

export { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyContent, EmptyMedia };
