import { LogOut, Settings } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Avatar, AvatarFallback } from "@components/UI/Avatar";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@components/UI/DropdownMenu";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@components/UI/Tooltip";
import { SettingsRoute } from "@constants/Routes";
import { useFlowPresent } from "@hooks/Flow";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { useSignOut } from "@hooks/SignOut";
import { UserInfo } from "@models/UserInfo";

export interface Props {
    userInfo?: UserInfo;
}

const AppBarItemAccountSettings = function (props: Props) {
    const { t: translate } = useTranslation();

    const navigate = useRouterNavigate();
    const doSignOut = useSignOut();
    const flowPresent = useFlowPresent();

    const handleSettingsClick = () => {
        navigate(SettingsRoute);
    };

    const handleSwitchUserClick = () => {
        doSignOut(true);
    };

    const handleLogoutClick = () => {
        doSignOut(false);
    };

    return props.userInfo ? (
        <div className="flex items-center text-center">
            <DropdownMenu>
                <TooltipProvider>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <DropdownMenuTrigger asChild>
                                <button
                                    id="account-menu"
                                    className="ml-4 rounded-full focus:outline-none focus:ring-2 focus:ring-ring"
                                >
                                    <Avatar className="size-8">
                                        <AvatarFallback>
                                            {props.userInfo.display_name.charAt(0).toUpperCase()}
                                        </AvatarFallback>
                                    </Avatar>
                                </button>
                            </DropdownMenuTrigger>
                        </TooltipTrigger>
                        <TooltipContent>{translate("Account Settings")}</TooltipContent>
                    </Tooltip>
                </TooltipProvider>
                <DropdownMenuContent align="end" sideOffset={8}>
                    <DropdownMenuItem id="account-menu-settings" onClick={handleSettingsClick}>
                        <Settings className="size-4" />
                        {translate("Settings")}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    {flowPresent ? (
                        <DropdownMenuItem id="account-menu-switch-user" onClick={handleSwitchUserClick}>
                            <LogOut className="size-4" />
                            {translate("Switch User")}
                        </DropdownMenuItem>
                    ) : null}
                    <DropdownMenuItem id="account-menu-logout" onClick={handleLogoutClick}>
                        <LogOut className="size-4" />
                        {translate("Logout")}
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>
        </div>
    ) : null;
};

export default AppBarItemAccountSettings;
