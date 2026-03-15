import { type ComponentProps, type Ref } from "react";

import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { cn } from "@utils/Styles";

interface OneTimeCodeInputProps extends ComponentProps<"input"> {
    label?: string;
    error?: boolean;
    inputRef?: Ref<HTMLInputElement>;
}

const OneTimeCodeTextField = function ({ className, error, inputRef, label, ...props }: OneTimeCodeInputProps) {
    return (
        <div className="flex flex-col gap-1.5">
            {label ? <Label>{label}</Label> : null}
            <Input
                {...props}
                ref={inputRef}
                className={cn(
                    "tracking-[0.5rem] text-center uppercase",
                    error && "border-destructive focus-visible:border-destructive focus-visible:ring-destructive/50",
                    className,
                )}
                spellCheck={false}
            />
        </div>
    );
};

export default OneTimeCodeTextField;
