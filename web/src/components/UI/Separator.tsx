import { type ComponentProps } from "react";

import { Separator as SeparatorPrimitive } from "radix-ui";

import { cn } from "@utils/Styles";

function Separator({
    className,
    decorative = true,
    orientation = "horizontal",
    ...props
}: ComponentProps<typeof SeparatorPrimitive.Root>) {
    return (
        <SeparatorPrimitive.Root
            data-slot="separator"
            decorative={decorative}
            orientation={orientation}
            className={cn(
                "shrink-0 bg-border data-[orientation=horizontal]:h-px data-[orientation=horizontal]:w-full data-[orientation=vertical]:h-full data-[orientation=vertical]:w-px",
                className,
            )}
            {...props}
        />
    );
}

export { Separator };
