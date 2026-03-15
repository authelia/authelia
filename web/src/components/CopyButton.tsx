import { ReactNode, useState } from "react";

import { Check, Copy } from "lucide-react";

import { Button } from "@components/UI/Button";
import { Spinner } from "@components/UI/Spinner";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { cn } from "@utils/Styles";

export interface Props {
    variant?: "contained" | "default" | "ghost" | "outline";
    tooltip: string;
    children: ReactNode;
    childrenCopied?: ReactNode;
    value: null | string;
    msTimeoutCopying?: number;
    msTimeoutCopied?: number;
    className?: string;
    fullWidth?: boolean;
}

const msTimeoutDefaultCopying = 500;
const msTimeoutDefaultCopied = 2000;

const CopyButton = function (props: Props) {
    const [isCopied, setIsCopied] = useState(false);
    const [isCopying, setIsCopying] = useState(false);
    const msTimeoutCopying = props.msTimeoutCopying ?? msTimeoutDefaultCopying;
    const msTimeoutCopied = props.msTimeoutCopied ?? msTimeoutDefaultCopied;

    const handleCopyToClipboard = () => {
        if (isCopied || !props.value || props.value === "") {
            return;
        }

        (async (value: string) => {
            setIsCopying(true);

            await navigator.clipboard.writeText(value);

            setTimeout(() => {
                setIsCopying(false);
                setIsCopied(true);
            }, msTimeoutCopying);

            setTimeout(() => {
                setIsCopied(false);
            }, msTimeoutCopied);
        })(props.value);
    };

    const variant = props.variant === "contained" ? "default" : (props.variant ?? "outline");
    const displayText = isCopied && props.childrenCopied ? props.childrenCopied : props.children;

    let icon;

    if (isCopying) {
        icon = <Spinner size={20} />;
    } else if (isCopied) {
        icon = <Check className="size-5" />;
    } else {
        icon = <Copy className="size-5" />;
    }

    const buttonClassName = cn(
        isCopied && "border-green-500 text-green-600",
        props.fullWidth && "w-full",
        props.className,
    );

    return props.value === null || props.value === "" ? (
        <Button variant={variant} disabled className={buttonClassName}>
            {icon}
            {displayText}
        </Button>
    ) : (
        <TooltipProvider>
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button
                        variant={variant}
                        onClick={isCopying ? undefined : handleCopyToClipboard}
                        className={buttonClassName}
                    >
                        {icon}
                        {displayText}
                    </Button>
                </TooltipTrigger>
                <TooltipContent>{props.tooltip}</TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
};

export default CopyButton;
