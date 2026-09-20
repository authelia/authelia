import { type ComponentProps } from "react";

import { Separator as SeparatorPrimitive } from "@base-ui/react/separator";

import { cn } from "@utils/Styles";

function Separator({ className, orientation = "horizontal", ...props }: ComponentProps<typeof SeparatorPrimitive>) {
    return (
        <SeparatorPrimitive
            data-slot="separator"
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
