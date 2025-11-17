import React, { useCallback } from "react";

import { Button, Tooltip } from "@mui/material";
import { useTranslation } from "react-i18next";

import { useSignOut } from "@hooks/SignOut";

export interface Props {
    id: string;
    text: string;
    tooltip?: string;
    preserve?: boolean;
}

const SignOutButton = function (props: Props) {
    const { t: translate } = useTranslation(["portal"]);

    const doSignOut = useSignOut();

    const handleSignOutClick = useCallback(() => {
        doSignOut(props.preserve ?? false);
    }, [doSignOut, props.preserve]);

    return props.tooltip ? (
        <Tooltip title={props.tooltip}>
            <Button id={props.id} color={"secondary"} onClick={handleSignOutClick}>
                {translate(props.text)}
            </Button>
        </Tooltip>
    ) : (
        <Button id={props.id} color={"secondary"} onClick={handleSignOutClick}>
            {translate(props.text)}
        </Button>
    );
};

export default SignOutButton;
