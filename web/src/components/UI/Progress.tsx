import { type ComponentProps } from "react";

import { Progress as ProgressPrimitive } from "@base-ui/react/progress";

import { cn } from "@utils/Styles";

function Progress({
    className,
    value,
    ...props
}: Omit<ComponentProps<typeof ProgressPrimitive.Root>, "value"> & { value?: null | number }) {
    const percentage = (100 * (value ?? 0)) / (props.max ?? 100);

    return (
        <ProgressPrimitive.Root data-slot="progress" value={value ?? null} {...props}>
            <ProgressPrimitive.Track
                data-slot="progress-track"
                className={cn("relative h-2 w-full overflow-hidden rounded-full bg-primary/20", className)}
            >
                <ProgressPrimitive.Indicator
                    data-slot="progress-indicator"
                    className="h-full w-full flex-1 bg-primary transition-all"
                    style={{ transform: `translateX(-${100 - percentage}%)` }}
                />
            </ProgressPrimitive.Track>
        </ProgressPrimitive.Root>
    );
}

export { Progress };
