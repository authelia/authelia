import { Fragment, JSX, ReactNode } from "react";

import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";

export interface Props {
    render: boolean;
    title: ReactNode;
    children: ReactNode;
    placement?: "bottom" | "left" | "right" | "top";
}

const ComponentWithTooltip = function (props: Props): JSX.Element {
    return (
        <Fragment>
            {props.render ? (
                <TooltipProvider>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <span>{props.children}</span>
                        </TooltipTrigger>
                        <TooltipContent side={props.placement}>{props.title}</TooltipContent>
                    </Tooltip>
                </TooltipProvider>
            ) : (
                props.children
            )}
        </Fragment>
    );
};

export default ComponentWithTooltip;
