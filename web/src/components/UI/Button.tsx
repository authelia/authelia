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
                default: "",
                ghost: "",
                outline: "border",
            },
        },
    },
);

type ButtonColor = "default" | "destructive" | "primary" | "secondary" | "success";

const colorClasses: Record<ButtonColor, { default: string; ghost: string; outline: string }> = {
    default: {
        default: "bg-primary text-primary-foreground hover:bg-primary/70",
        ghost: "hover:bg-accent/10",
        outline:
            "border-input bg-background shadow-xs hover:bg-accent hover:text-accent-foreground dark:border-input dark:bg-input/30 dark:hover:bg-input/50",
    },
    destructive: {
        default: "bg-destructive text-white hover:bg-destructive/70",
        ghost: "text-destructive hover:bg-destructive/10",
        outline: "border-destructive text-destructive hover:bg-destructive/10",
    },
    primary: {
        default: "bg-primary text-primary-foreground hover:bg-primary/70",
        ghost: "text-primary hover:bg-primary/10",
        outline: "border-primary text-primary hover:bg-primary/10",
    },
    secondary: {
        default: "bg-secondary text-secondary-foreground hover:bg-secondary/70",
        ghost: "text-secondary hover:bg-secondary/10",
        outline: "border-secondary text-secondary hover:bg-secondary/10",
    },
    success: {
        default: "bg-green-600 text-white hover:bg-green-700",
        ghost: "text-green-600 hover:bg-green-600/10 dark:text-green-500 dark:hover:bg-green-500/10",
        outline:
            "border-green-600 text-green-600 hover:bg-green-600/10 dark:border-green-500 dark:text-green-500 dark:hover:bg-green-500/10",
    },
};

function getColorClasses(variant: string = "default", color: ButtonColor = "default"): string {
    const entry = colorClasses[color] ?? colorClasses.default;

    return entry[variant as keyof typeof entry] ?? entry.default;
}

function Button({
    asChild = false,
    className,
    color,
    size = "default",
    variant = "default",
    ...props
}: ComponentProps<"button"> &
    VariantProps<typeof buttonVariants> & {
        asChild?: boolean;
        color?: ButtonColor;
    }) {
    const Comp = asChild ? Slot.Root : "button";

    return (
        <Comp
            data-slot="button"
            data-variant={variant}
            data-size={size}
            className={cn(buttonVariants({ size, variant }), getColorClasses(variant ?? undefined, color), className)}
            {...props}
        />
    );
}

export { Button, buttonVariants };
export type { ButtonColor };
