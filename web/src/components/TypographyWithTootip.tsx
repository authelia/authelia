import React, { Fragment } from "react";

import { Tooltip, Typography } from "@mui/material";
import { Variant } from "@mui/material/styles/createTypography";
import { CSSProperties } from "@mui/styles";

export interface Props {
    variant: Variant;

    value?: string;
    style?: CSSProperties;

    tooltip?: string;
    tooltipStyle?: CSSProperties;
}

export default function TypographyWithTooltip(props: Props): JSX.Element {
    return (
        <Fragment>
            {props.tooltip ? (
                <Tooltip title={props.tooltip} style={props.tooltipStyle}>
                    <Typography variant={props.variant} style={props.style}>
                        {props.value}
                    </Typography>
                </Tooltip>
            ) : (
                <Typography variant={props.variant} style={props.style}>
                    {props.value}
                </Typography>
            )}
        </Fragment>
    );
}
