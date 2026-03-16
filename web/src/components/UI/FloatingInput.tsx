import { type ComponentProps, forwardRef, useId, useState } from "react";

import { cn } from "@utils/Styles";

interface FloatingInputProps extends ComponentProps<"input"> {
    label: string;
    error?: boolean;
}

const FloatingInput = forwardRef<HTMLInputElement, FloatingInputProps>(
    ({ className, error, label, onBlur, onFocus, ...props }, ref) => {
        const id = useId();
        const inputId = props.id ?? id;
        const [focused, setFocused] = useState(false);
        const hasValue = props.value !== undefined && props.value !== "";
        const isFloating = focused || hasValue;

        return (
            <div className="relative w-full">
                <input
                    ref={ref}
                    id={inputId}
                    aria-invalid={error}
                    data-slot="floating-input"
                    className={cn(
                        "peer h-14 w-full rounded border border-input bg-transparent px-3 pt-5 pb-1.5 text-base outline-none transition-colors",
                        "focus:border-primary focus:ring-2 focus:ring-primary/30",
                        "disabled:cursor-not-allowed disabled:opacity-50",
                        "placeholder-transparent",
                        error && "border-destructive focus:border-destructive focus:ring-destructive/30",
                        className,
                    )}
                    placeholder={label}
                    onFocus={(e) => {
                        setFocused(true);
                        onFocus?.(e);
                    }}
                    onBlur={(e) => {
                        setFocused(false);
                        onBlur?.(e);
                    }}
                    {...props}
                />
                <label
                    htmlFor={inputId}
                    className={cn(
                        "pointer-events-none absolute left-3 transition-all duration-200",
                        "text-muted-foreground",
                        isFloating ? "top-1 text-xs" : "top-1/2 -translate-y-1/2 text-base",
                        focused && !error && "text-primary",
                        error && "text-destructive",
                    )}
                >
                    {label}
                </label>
            </div>
        );
    },
);

FloatingInput.displayName = "FloatingInput";

export { FloatingInput };
