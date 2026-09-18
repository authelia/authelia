"use client";

import { type ComponentProps } from "react";

import { Popover as PopoverPrimitive } from "@base-ui/react/popover";

import { cn } from "@utils/Styles";

function Popover({ ...props }: ComponentProps<typeof PopoverPrimitive.Root>) {
    return <PopoverPrimitive.Root {...props} />;
}

function PopoverTrigger({ ...props }: ComponentProps<typeof PopoverPrimitive.Trigger>) {
    return <PopoverPrimitive.Trigger data-slot="popover-trigger" {...props} />;
}

function PopoverClose({ ...props }: ComponentProps<typeof PopoverPrimitive.Close>) {
    return <PopoverPrimitive.Close data-slot="popover-close" {...props} />;
}

function PopoverContent({
    align,
    children,
    className,
    side,
    sideOffset = 4,
    ...props
}: ComponentProps<typeof PopoverPrimitive.Popup> & {
    align?: ComponentProps<typeof PopoverPrimitive.Positioner>["align"];
    side?: ComponentProps<typeof PopoverPrimitive.Positioner>["side"];
    sideOffset?: number;
}) {
    return (
        <PopoverPrimitive.Portal>
            <PopoverPrimitive.Positioner align={align} side={side} sideOffset={sideOffset} className="z-50">
                <PopoverPrimitive.Popup
                    data-slot="popover-content"
                    className={cn(
                        "w-72 rounded-md border bg-popover p-3 text-popover-foreground shadow-md outline-none",
                        "origin-(--transform-origin) transition-[opacity,transform] duration-150",
                        "data-[starting-style]:scale-95 data-[starting-style]:opacity-0",
                        "data-[ending-style]:scale-95 data-[ending-style]:opacity-0",
                        className,
                    )}
                    {...props}
                >
                    {children}
                </PopoverPrimitive.Popup>
            </PopoverPrimitive.Positioner>
        </PopoverPrimitive.Portal>
    );
}

export { Popover, PopoverClose, PopoverContent, PopoverTrigger };
