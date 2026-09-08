"use client";

import { type ComponentProps } from "react";

import { Tooltip as TooltipPrimitive } from "@base-ui/react/tooltip";

import { cn } from "@utils/Styles";

function TooltipProvider({ delay = 0, ...props }: ComponentProps<typeof TooltipPrimitive.Provider>) {
    return <TooltipPrimitive.Provider delay={delay} {...props} />;
}

function Tooltip({ ...props }: ComponentProps<typeof TooltipPrimitive.Root>) {
    return <TooltipPrimitive.Root {...props} />;
}

function TooltipTrigger({ ...props }: ComponentProps<typeof TooltipPrimitive.Trigger>) {
    return <TooltipPrimitive.Trigger data-slot="tooltip-trigger" {...props} />;
}

function TooltipContent({
    children,
    className,
    side,
    sideOffset = 4,
    ...props
}: ComponentProps<typeof TooltipPrimitive.Popup> & {
    side?: ComponentProps<typeof TooltipPrimitive.Positioner>["side"];
    sideOffset?: number;
}) {
    return (
        <TooltipPrimitive.Portal>
            <TooltipPrimitive.Positioner side={side} sideOffset={sideOffset} className="z-50 pointer-events-none">
                <TooltipPrimitive.Popup
                    data-slot="tooltip-content"
                    className={cn(
                        "w-fit rounded-md bg-foreground px-3 py-1.5 text-xs text-balance text-background",
                        "origin-(--transform-origin) transition-[opacity,transform] duration-150",
                        "data-[starting-style]:scale-95 data-[starting-style]:opacity-0",
                        "data-[ending-style]:scale-95 data-[ending-style]:opacity-0",
                        className,
                    )}
                    {...props}
                >
                    {children}
                    <TooltipPrimitive.Arrow className="size-2.5 -translate-y-1/2 rotate-45 rounded-[2px] bg-foreground" />
                </TooltipPrimitive.Popup>
            </TooltipPrimitive.Positioner>
        </TooltipPrimitive.Portal>
    );
}

export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider };
