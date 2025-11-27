import { ReactNode, useState } from "react";

import { Check, ContentCopy } from "@mui/icons-material";
import { Button, CircularProgress, SxProps, Tooltip } from "@mui/material";

export interface Props {
    variant?: "contained" | "outlined" | "text";
    tooltip: string;
    children: ReactNode;
    childrenCopied?: ReactNode;
    value: null | string;
    msTimeoutCopying?: number;
    msTimeoutCopied?: number;
    sx?: SxProps;
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

    const variant = props.variant ?? "outlined";
    const color = isCopied ? "success" : "primary";
    const displayText = isCopied && props.childrenCopied ? props.childrenCopied : props.children;

    let icon;

    if (isCopying) {
        icon = <CircularProgress color="inherit" size={20} />;
    } else if (isCopied) {
        icon = <Check />;
    } else {
        icon = <ContentCopy />;
    }

    return props.value === null || props.value === "" ? (
        <Button variant={variant} color={color} disabled sx={props.sx} fullWidth={props.fullWidth} startIcon={icon}>
            {displayText}
        </Button>
    ) : (
        <Tooltip title={props.tooltip}>
            <Button
                variant={variant}
                color={color}
                onClick={isCopying ? undefined : handleCopyToClipboard}
                sx={props.sx}
                fullWidth={props.fullWidth}
                startIcon={icon}
            >
                {displayText}
            </Button>
        </Tooltip>
    );
};

export default CopyButton;
