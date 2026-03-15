import { FC } from "react";

import { useTranslation } from "react-i18next";

import { ScopeAvatar } from "@components/OpenIDConnect";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { formatScope } from "@services/ConsentOpenIDConnect";

export interface Props {
    scopes: string[];
}

const DecisionFormScopes: FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent"]);

    return (
        <div className="w-full">
            <div className="text-center">
                <ul className="my-4 inline-block list-none bg-card">
                    {props.scopes.map((scope: string) => (
                        <TooltipProvider key={scope}>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <li id={"scope-" + scope} className="flex items-center gap-3 px-2 py-1">
                                        <span className="flex-shrink-0">{ScopeAvatar(scope)}</span>
                                        <span>{formatScope(translate(`scopes.${scope}`), scope)}</span>
                                    </li>
                                </TooltipTrigger>
                                <TooltipContent>{translate("Scope", { name: scope })}</TooltipContent>
                            </Tooltip>
                        </TooltipProvider>
                    ))}
                </ul>
            </div>
        </div>
    );
};

export default DecisionFormScopes;
