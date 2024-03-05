import React, { Fragment } from "react";

import { Tooltip, Typography } from "@mui/material";
import { Variant } from "@mui/material/styles/createTypography";

export interface Props {
    variant: Variant;

    value?: string;

    tooltip?: string;
}

const TypographyWithTooltip = function (props: Props): JSX.Element {
    return (
        <Fragment>
            {props.tooltip ? (
                <Tooltip title={props.tooltip}>
                    <Typography variant={props.variant}>{props.value}</Typography>
                </Tooltip>
            ) : (
                <Typography variant={props.variant}>{props.value}</Typography>
            )}
        </Fragment>
    );
};

export default TypographyWithTooltip;
