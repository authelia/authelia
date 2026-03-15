import { Fragment, JSX } from "react";

import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { cn } from "@utils/Styles";

export type TypographyVariant = "body1" | "body2" | "h1" | "h2" | "h3" | "h4" | "h5" | "h6" | "subtitle1" | "subtitle2";

export interface Props {
    variant: TypographyVariant;

    value?: string;

    tooltip?: string;
}

const variantClassMap: Record<TypographyVariant, { tag: keyof JSX.IntrinsicElements; className: string }> = {
    body1: { className: "text-base", tag: "p" },
    body2: { className: "text-sm", tag: "p" },
    h1: { className: "text-5xl font-light", tag: "h1" },
    h2: { className: "text-4xl font-light", tag: "h2" },
    h3: { className: "text-3xl", tag: "h3" },
    h4: { className: "text-2xl", tag: "h4" },
    h5: { className: "text-2xl font-normal", tag: "h5" },
    h6: { className: "text-xl font-medium", tag: "h6" },
    subtitle1: { className: "text-base", tag: "h6" },
    subtitle2: { className: "text-sm font-medium", tag: "h6" },
};

const TypographyWithTooltip = function (props: Props): JSX.Element {
    const { className, tag: Tag } = variantClassMap[props.variant];

    const typography = <Tag className={cn(className)}>{props.value}</Tag>;

    return (
        <Fragment>
            {props.tooltip ? (
                <TooltipProvider>
                    <Tooltip>
                        <TooltipTrigger asChild>{typography}</TooltipTrigger>
                        <TooltipContent>{props.tooltip}</TooltipContent>
                    </Tooltip>
                </TooltipProvider>
            ) : (
                typography
            )}
        </Fragment>
    );
};

export default TypographyWithTooltip;
