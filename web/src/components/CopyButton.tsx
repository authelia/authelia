import React, { ReactNode, useState } from "react";

import { Check, ContentCopy } from "@mui/icons-material";
import { Button, CircularProgress, SxProps, Tooltip } from "@mui/material";

export interface Props {
    variant?: "contained" | "outlined" | "text";
    tooltip: string;
    children: ReactNode;
    childrenCopied?: ReactNode;
    value: null | string;
    xs?: number;
    msTimeoutCopying?: number;
    msTimeoutCopied?: number;
    sx?: SxProps;
    fullWidth?: boolean;
}

const msTmeoutDefaultCopying = 500;
const msTmeoutDefaultCopied = 2000;

const CopyButton = function (props: Props) {
    const [isCopied, setIsCopied] = useState(false);
    const [isCopying, setIsCopying] = useState(false);
    const msTimeoutCopying = props.msTimeoutCopying ? props.msTimeoutCopying : msTmeoutDefaultCopying;
    const msTimeoutCopied = props.msTimeoutCopied ? props.msTimeoutCopied : msTmeoutDefaultCopied;

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

    return props.value === null || props.value === "" ? (
        <Button
            variant={props.variant ? props.variant : "outlined"}
            color={isCopied ? "success" : "primary"}
            disabled
            sx={props.sx}
            fullWidth={props.fullWidth}
            startIcon={
                isCopying ? <CircularProgress color="inherit" size={20} /> : isCopied ? <Check /> : <ContentCopy />
            }
        >
            {isCopied && props.childrenCopied ? props.childrenCopied : props.children}
        </Button>
    ) : (
        <Tooltip title={props.tooltip}>
            <Button
                variant={props.variant ? props.variant : "outlined"}
                color={isCopied ? "success" : "primary"}
                onClick={isCopying ? undefined : handleCopyToClipboard}
                sx={props.sx}
                fullWidth={props.fullWidth}
                startIcon={
                    isCopying ? <CircularProgress color="inherit" size={20} /> : isCopied ? <Check /> : <ContentCopy />
                }
            >
                {isCopied && props.childrenCopied ? props.childrenCopied : props.children}
            </Button>
        </Tooltip>
    );
};

export default CopyButton;
