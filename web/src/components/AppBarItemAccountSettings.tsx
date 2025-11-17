import React, { Fragment, useState } from "react";

import { Logout, Settings } from "@mui/icons-material";
import { Avatar, Box, Divider, IconButton, ListItemIcon, Menu, MenuItem, Tooltip } from "@mui/material";
import { useTranslation } from "react-i18next";

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

    const [elementAccountSettings, setElementAccountSettings] = useState<HTMLElement | null>(null);

    const navigate = useRouterNavigate();
    const doSignOut = useSignOut();
    const flowPresent = useFlowPresent();

    const handleSettingsClick = () => {
        handleAccountSettingsClose();

        navigate(SettingsRoute);
    };

    const handleSwitchUserClick = () => {
        handleAccountSettingsClose();

        doSignOut(true);
    };

    const handleLogoutClick = () => {
        handleAccountSettingsClose();

        doSignOut(false);
    };

    const open = Boolean(elementAccountSettings);

    const handleAccountSettingsClick = (event: React.MouseEvent<HTMLElement>) => {
        setElementAccountSettings(event.currentTarget);
    };

    const handleAccountSettingsClose = () => {
        setElementAccountSettings(null);
    };

    return props.userInfo ? (
        <Fragment>
            <Box sx={{ alignItems: "center", display: "flex", textAlign: "center" }}>
                <Tooltip title={translate("Account Settings")}>
                    <IconButton
                        id={"account-menu"}
                        onClick={handleAccountSettingsClick}
                        size={"small"}
                        sx={{ ml: 2 }}
                        aria-controls={open ? "account-menu" : undefined}
                        aria-haspopup={"true"}
                        aria-expanded={open ? "true" : undefined}
                    >
                        <Avatar sx={{ height: 32, width: 32 }}>
                            {props.userInfo.display_name.charAt(0).toUpperCase()}
                        </Avatar>
                    </IconButton>
                </Tooltip>
            </Box>
            <Menu
                anchorEl={elementAccountSettings}
                id={"account-menu"}
                open={open}
                onClose={handleAccountSettingsClose}
                onClick={handleAccountSettingsClose}
                slotProps={{
                    paper: {
                        elevation: 0,
                        sx: {
                            "&:before": {
                                bgcolor: "background.paper",
                                content: '""',
                                display: "block",
                                height: 10,
                                position: "absolute",
                                right: 14,
                                top: 0,
                                transform: "translateY(-50%) rotate(45deg)",
                                width: 10,
                                zIndex: 0,
                            },
                            "& .MuiAvatar-root": {
                                height: 32,
                                ml: -0.5,
                                mr: 1,
                                width: 32,
                            },
                            filter: "drop-shadow(0px 2px 8px rgba(0,0,0,0.32))",
                            mt: 1.5,
                            overflow: "visible",
                        },
                    },
                }}
                transformOrigin={{ horizontal: "right", vertical: "top" }}
                anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
            >
                <MenuItem onClick={handleSettingsClick} id={"account-menu-settings"}>
                    <ListItemIcon>
                        <Settings fontSize="small" />
                    </ListItemIcon>
                    {translate("Settings")}
                </MenuItem>
                <Divider />
                {flowPresent ? (
                    <MenuItem onClick={handleSwitchUserClick} id={"account-menu-switch-user"}>
                        <ListItemIcon>
                            <Logout fontSize="small" />
                        </ListItemIcon>
                        {translate("Switch User")}
                    </MenuItem>
                ) : null}
                <MenuItem onClick={handleLogoutClick} id={"account-menu-logout"}>
                    <ListItemIcon>
                        <Logout fontSize="small" />
                    </ListItemIcon>
                    {translate("Logout")}
                </MenuItem>
            </Menu>
        </Fragment>
    ) : null;
};

export default AppBarItemAccountSettings;
