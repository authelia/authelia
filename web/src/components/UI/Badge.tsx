import { type ComponentProps } from "react";

import { useRender } from "@base-ui/react/use-render";
import { type VariantProps, cva } from "class-variance-authority";

import { cn } from "@utils/Styles";

const badgeVariants = cva(
    "inline-flex w-fit shrink-0 items-center justify-center gap-1 overflow-hidden rounded-md border px-2 py-0.5 text-xs font-medium whitespace-nowrap transition-[color,box-shadow] focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 [&_svg]:pointer-events-none [&_svg]:size-3",
    {
        defaultVariants: {
            variant: "default",
        },
        variants: {
            variant: {
                default: "border-transparent bg-primary text-primary-foreground [a&]:hover:bg-primary/90",
                destructive:
                    "border-transparent bg-destructive text-white [a&]:hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:bg-destructive/70 dark:focus-visible:ring-destructive/40",
                outline: "text-foreground [a&]:hover:bg-accent [a&]:hover:text-accent-foreground",
                secondary: "border-transparent bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/90",
            },
        },
    },
);

function Badge({
    className,
    render,
    variant = "default",
    ...props
}: ComponentProps<"span"> &
    VariantProps<typeof badgeVariants> & {
        render?: useRender.RenderProp;
    }) {
    return useRender({
        props: {
            className: cn(badgeVariants({ variant }), className),
            "data-slot": "badge",
            ...props,
        },
        render: render ?? <span />,
    });
}

export { Badge, badgeVariants };
