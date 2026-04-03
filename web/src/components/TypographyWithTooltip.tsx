import { Fragment, JSX } from "react";

import { Tooltip, Typography } from "@mui/material";
import { TypographyVariant } from "@mui/material/styles";

export interface Props {
    variant: TypographyVariant;

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
