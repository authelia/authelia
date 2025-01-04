import React, { Fragment } from "react";

import { Tooltip } from "@mui/material";
import { TooltipProps } from "@mui/material/Tooltip";

export interface Props extends TooltipProps {
    render: boolean;
}

interface ComponentProps extends Omit<Props, "render"> {}

const ComponentWithTooltip = function (props: Props): React.JSX.Element {
    const tooltipProps = props as ComponentProps;

    return (
        <Fragment>
            {props.render ? (
                <Tooltip {...tooltipProps}>
                    <span>{props.children}</span>
                </Tooltip>
            ) : (
                props.children
            )}
        </Fragment>
    );
};

export default ComponentWithTooltip;
