import { type ComponentProps } from "react";

import { type VariantProps, cva } from "class-variance-authority";
import { Slot } from "radix-ui";

import { cn } from "@utils/Styles";

const buttonVariants = cva(
    "inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-medium uppercase whitespace-nowrap transition-all outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:pointer-events-none disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
    {
        defaultVariants: {
            size: "default",
            variant: "default",
        },
        variants: {
            size: {
                default: "h-10 px-4 py-2 has-[>svg]:px-3",
                icon: "size-9",
                "icon-lg": "size-10",
                "icon-sm": "size-8",
                "icon-xs": "size-6 rounded-md [&_svg:not([class*='size-'])]:size-3",
                lg: "h-10 rounded-md px-6 has-[>svg]:px-4",
                sm: "h-8 gap-1.5 rounded-md px-3 has-[>svg]:px-2.5",
                xs: "h-6 gap-1 rounded-md px-2 text-xs has-[>svg]:px-1.5 [&_svg:not([class*='size-'])]:size-3",
            },
            variant: {
                default: "bg-primary text-primary-foreground hover:bg-primary/90",
                destructive:
                    "bg-destructive text-white hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:bg-destructive/60 dark:focus-visible:ring-destructive/40",
                ghost: "hover:bg-secondary/10 hover:text-secondary dark:hover:bg-secondary/10",
                link: "text-primary underline-offset-4 hover:underline",
                outline:
                    "border bg-background shadow-xs hover:bg-accent hover:text-accent-foreground dark:border-input dark:bg-input/30 dark:hover:bg-input/50",
                secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
            },
        },
    },
);

function Button({
    asChild = false,
    className,
    size = "default",
    variant = "default",
    ...props
}: ComponentProps<"button"> &
    VariantProps<typeof buttonVariants> & {
        asChild?: boolean;
    }) {
    const Comp = asChild ? Slot.Root : "button";

    return (
        <Comp
            data-slot="button"
            data-variant={variant}
            data-size={size}
            className={cn(buttonVariants({ className, size, variant }))}
            {...props}
        />
    );
}

export { Button, buttonVariants };
