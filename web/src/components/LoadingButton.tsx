import React from "react";

import { Button, CircularProgress } from "@mui/material";
import { ButtonProps } from "@mui/material/Button";

export interface Props extends ButtonProps {
    loading: boolean;
}

function LoadingButton(props: Props) {
    let { loading, ...childProps } = props;
    if (loading) {
        childProps = {
            ...childProps,
            startIcon: <CircularProgress color="inherit" size={20} />,
            color: "inherit",
            onClick: undefined,
        };
    }
    return <Button {...childProps}></Button>;
}

export default LoadingButton;
