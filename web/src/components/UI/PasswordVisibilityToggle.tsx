import { ComponentProps } from "react";

import { Eye, EyeOff } from "lucide-react";

interface Props extends ComponentProps<"button"> {
    label: string;
    showPassword: boolean;
}

function PasswordVisibilityToggle({ label, showPassword, ...props }: Props) {
    return (
        <button
            type="button"
            className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-muted-foreground hover:text-foreground"
            aria-label={label}
            {...props}
        >
            {showPassword ? <Eye className="h-5 w-5" /> : <EyeOff className="h-5 w-5" />}
        </button>
    );
}

export { PasswordVisibilityToggle };
