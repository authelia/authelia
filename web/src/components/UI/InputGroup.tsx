"use client";

import { type ComponentProps } from "react";

import { type VariantProps, cva } from "class-variance-authority";

import { Button } from "@components/UI/Button";
import { Input } from "@components/UI/Input";
import { cn } from "@utils/Styles";

function InputGroup({ className, ...props }: ComponentProps<"div">) {
    return (
        <div
            data-slot="input-group"
            role="group"
            className={cn(
                "group/input-group relative flex w-full items-center rounded-md border border-input shadow-xs transition-[color,box-shadow] outline-none dark:bg-input/30",
                "h-14 min-w-0",

                // Variants based on alignment.
                "has-[>[data-align=inline-start]]:[&>input]:pl-2",
                "has-[>[data-align=inline-end]]:[&>input]:pr-2",
                "has-[>[data-align=block-start]]:h-auto has-[>[data-align=block-start]]:flex-col has-[>[data-align=block-start]]:[&>input]:pb-3",
                "has-[>[data-align=block-end]]:h-auto has-[>[data-align=block-end]]:flex-col has-[>[data-align=block-end]]:[&>input]:pt-3",

                // Focus state.
                "has-[[data-slot=input-group-control]:focus-visible]:border-ring has-[[data-slot=input-group-control]:focus-visible]:ring-[3px] has-[[data-slot=input-group-control]:focus-visible]:ring-ring/50",

                // Error state.
                "has-[[data-slot][aria-invalid=true]]:border-destructive has-[[data-slot][aria-invalid=true]]:ring-destructive/20 dark:has-[[data-slot][aria-invalid=true]]:ring-destructive/40",

                className,
            )}
            {...props}
        />
    );
}

const inputGroupAddonVariants = cva(
    "flex h-auto cursor-text items-center justify-center gap-2 py-1.5 text-sm font-medium text-muted-foreground select-none group-data-[disabled=true]/input-group:opacity-50 [&>svg:not([class*='size-'])]:size-4",
    {
        defaultVariants: {
            align: "inline-start",
        },
        variants: {
            align: {
                "block-end":
                    "order-last w-full justify-start px-3 pb-3 group-has-[>input]/input-group:pb-2.5 [.border-t]:pt-3",
                "block-start":
                    "order-first w-full justify-start px-3 pt-3 group-has-[>input]/input-group:pt-2.5 [.border-b]:pb-3",
                "inline-end": "order-last pr-3 has-[>button]:mr-[-0.45rem]",
                "inline-start": "order-first pl-3 has-[>button]:ml-[-0.45rem]",
            },
        },
    },
);

function InputGroupAddon({
    align = "inline-start",
    className,
    ...props
}: ComponentProps<"div"> & VariantProps<typeof inputGroupAddonVariants>) {
    return (
        <div
            role="group"
            data-slot="input-group-addon"
            data-align={align}
            className={cn(inputGroupAddonVariants({ align }), className)}
            onClick={(event) => {
                if ((event.target as HTMLElement).closest("button")) {
                    return;
                }

                event.currentTarget.parentElement?.querySelector("input")?.focus();
            }}
            {...props}
        />
    );
}

const inputGroupButtonVariants = cva("flex items-center gap-2 text-sm shadow-none", {
    defaultVariants: {
        size: "xs",
    },
    variants: {
        size: {
            "icon-sm": "size-8 p-0 has-[>svg]:p-0",
            "icon-xs": "size-6 rounded-[calc(var(--radius)-5px)] p-0 has-[>svg]:p-0",
            sm: "h-8 gap-1.5 rounded-md px-2.5 has-[>svg]:px-2.5",
            xs: "h-6 gap-1 rounded-[calc(var(--radius)-5px)] px-2 has-[>svg]:px-2 [&>svg:not([class*='size-'])]:size-3.5",
        },
    },
});

function InputGroupButton({
    className,
    size = "xs",
    type = "button",
    variant = "ghost",
    ...props
}: Omit<ComponentProps<typeof Button>, "size"> & VariantProps<typeof inputGroupButtonVariants>) {
    return (
        <Button
            type={type}
            data-size={size}
            variant={variant}
            className={cn(inputGroupButtonVariants({ size }), className)}
            {...props}
        />
    );
}

function InputGroupText({ className, ...props }: ComponentProps<"span">) {
    return (
        <span
            className={cn(
                "flex items-center gap-2 text-sm text-muted-foreground [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4",
                className,
            )}
            {...props}
        />
    );
}

function InputGroupInput({ className, ...props }: ComponentProps<typeof Input>) {
    return (
        <Input
            data-slot="input-group-control"
            className={cn(
                "h-full flex-1 rounded-none border-0 bg-transparent shadow-none focus-visible:ring-0 dark:bg-transparent",
                className,
            )}
            {...props}
        />
    );
}

export { InputGroup, InputGroupAddon, InputGroupButton, InputGroupText, InputGroupInput };
