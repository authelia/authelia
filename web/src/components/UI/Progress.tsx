import { type ComponentProps } from "react";

import { Progress as ProgressPrimitive } from "@base-ui/react/progress";

import { cn } from "@utils/Styles";

function Progress({
    className,
    max,
    min,
    value,
    ...props
}: Omit<ComponentProps<typeof ProgressPrimitive.Root>, "value"> & { value?: null | number }) {
    const start = Number.isFinite(min) ? (min as number) : 0;
    const end = Number.isFinite(max) && (max as number) > start ? (max as number) : start + 100;
    const current = Number.isFinite(value) ? Math.min(Math.max(value as number, start), end) : null;
    const percentage = current === null ? 0 : (100 * (current - start)) / (end - start);

    return (
        <ProgressPrimitive.Root data-slot="progress" min={start} max={end} value={current} {...props}>
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
