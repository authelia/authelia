import { MouseEvent, ReactNode } from "react";

import { AlertTriangle, Info, Pencil, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Card } from "@components/UI/Card";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { useRelativeTime } from "@hooks/RelativeTimeString";

interface Props {
    id: string;
    icon?: ReactNode;
    description: string;
    qualifier: string;
    problem?: boolean;
    created_at: Date;
    last_used_at?: Date;
    tooltipInformation?: string;
    tooltipInformationProblem?: string;
    tooltipEdit?: string;
    tooltipDelete: string;
    handleInformation?: (_event: MouseEvent<HTMLElement>) => void;
    handleEdit?: (_event: MouseEvent<HTMLElement>) => void;
    handleDelete: (_event: MouseEvent<HTMLElement>) => void;
}

const CredentialItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const timeSinceAdded = useRelativeTime(props.created_at);
    const timeSinceLastUsed = useRelativeTime(props.last_used_at || new Date(0));

    return (
        <Card id={props.id} className="p-0">
            <div className="p-6">
                <div className="flex items-center h-full w-full">
                    <div className="shrink-0 mr-2 md:mr-4 xl:mr-6">{props.icon}</div>
                    <div className="flex-1 min-w-0">
                        <div className="flex flex-col">
                            <div className="flex flex-row items-center">
                                <span id={`${props.id}-description`} className="font-bold inline">
                                    {props.description}
                                </span>
                                <span className="hidden sm:inline text-sm px-4">{props.qualifier}</span>
                            </div>
                            <span className="hidden sm:block text-xs text-muted-foreground">
                                {`${translate("Added")} ${timeSinceAdded}`}
                            </span>
                            <span className="hidden sm:block text-xs text-muted-foreground">
                                {props.last_used_at === undefined
                                    ? translate("Never used")
                                    : `${translate("Last Used")} ${timeSinceLastUsed}`}
                            </span>
                        </div>
                    </div>
                    <div className="flex items-center justify-end gap-1">
                        {props.handleInformation ? (
                            <TooltipElement
                                tooltip={props.problem ? props.tooltipInformationProblem : props.tooltipInformation}
                            >
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={props.handleInformation}
                                    id={`${props.id}-information`}
                                >
                                    {props.problem ? <AlertTriangle className="text-amber-500" /> : <Info />}
                                </Button>
                            </TooltipElement>
                        ) : null}
                        {props.handleEdit ? (
                            <TooltipElement tooltip={props.tooltipEdit}>
                                <Button variant="ghost" size="icon" onClick={props.handleEdit} id={`${props.id}-edit`}>
                                    <Pencil />
                                </Button>
                            </TooltipElement>
                        ) : null}
                        <TooltipProvider>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        onClick={props.handleDelete}
                                        id={`${props.id}-delete`}
                                    >
                                        <Trash2 />
                                    </Button>
                                </TooltipTrigger>
                                <TooltipContent>{props.tooltipDelete}</TooltipContent>
                            </Tooltip>
                        </TooltipProvider>
                    </div>
                </div>
            </div>
        </Card>
    );
};

interface TooltipElementProps {
    tooltip?: string;
    children: ReactNode;
}

const TooltipElement = function (props: TooltipElementProps) {
    return props.tooltip !== undefined && props.tooltip !== "" ? (
        <TooltipProvider>
            <Tooltip>
                <TooltipTrigger asChild>{props.children}</TooltipTrigger>
                <TooltipContent>{props.tooltip}</TooltipContent>
            </Tooltip>
        </TooltipProvider>
    ) : (
        <>{props.children}</>
    );
};

export default CredentialItem;
