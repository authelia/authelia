import { Fragment, JSX } from "react";

import { Box, Tooltip } from "@mui/material";
import { TooltipProps } from "@mui/material/Tooltip";

export interface Props extends TooltipProps {
    render: boolean;
}

interface ComponentProps extends Omit<Props, "render"> {}

const ComponentWithTooltip = function (props: Props): JSX.Element {
    const tooltipProps = props as ComponentProps;

    return (
        <Fragment>
            {props.render ? (
                <Tooltip {...tooltipProps}>
                    <Box component={"span"}>{props.children}</Box>
                </Tooltip>
            ) : (
                props.children
            )}
        </Fragment>
    );
};

export default ComponentWithTooltip;
