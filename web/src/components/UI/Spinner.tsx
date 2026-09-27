import { type ComponentProps } from "react";

import { cn } from "cn";
import { Loader2 } from "lucide-react";

interface SpinnerProps extends ComponentProps<"div"> {
    size?: number;
    color?: string;
}

function Spinner({ className, color, size = 24, ...props }: Readonly<SpinnerProps>) {
    return (
        <div data-slot="spinner" className={cn("inline-flex items-center justify-center", className)} {...props}>
            <Loader2 className="animate-spin" style={{ color, height: size, width: size }} />
        </div>
    );
}

export { Spinner };
