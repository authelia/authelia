import { useCallback } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/UI/Tooltip";
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
        doSignOut(props.preserve ? props.preserve : false);
    }, [doSignOut, props.preserve]);

    return props.tooltip ? (
        <Tooltip>
            <TooltipTrigger asChild>
                <Button
                    id={props.id}
                    variant={"ghost"}
                    className="text-sm tracking-wide"
                    color={"secondary"}
                    onClick={handleSignOutClick}
                >
                    {translate(props.text)}
                </Button>
            </TooltipTrigger>
            <TooltipContent>{props.tooltip}</TooltipContent>
        </Tooltip>
    ) : (
        <Button
            id={props.id}
            variant={"ghost"}
            className="text-sm tracking-wide"
            color={"secondary"}
            onClick={handleSignOutClick}
        >
            {translate(props.text)}
        </Button>
    );
};

export default SignOutButton;
