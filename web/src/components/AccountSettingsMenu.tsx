import React, { Fragment, useState } from "react";

import { Logout, Settings } from "@mui/icons-material";
import { Avatar, Box, Divider, IconButton, ListItemIcon, Menu, MenuItem, Tooltip } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { LogoutRoute, SettingsRoute } from "@constants/Routes";
import { UserInfo } from "@models/UserInfo";

export interface Props {
    userInfo: UserInfo;
}

const AccountSettingsMenu = function (props: Props) {
    const { t: translate } = useTranslation();

    const [elementAccountSettings, setElementAccountSettings] = useState<null | HTMLElement>(null);

    const navigate = useNavigate();

    const handleSettingsClick = () => {
        handleAccountSettingsClose();

        navigate({ pathname: SettingsRoute });
    };

    const handleLogoutClick = () => {
        handleAccountSettingsClose();

        navigate({ pathname: LogoutRoute });
    };

    const open = Boolean(elementAccountSettings);

    const handleAccountSettingsClick = (event: React.MouseEvent<HTMLElement>) => {
        setElementAccountSettings(event.currentTarget);
    };

    const handleAccountSettingsClose = () => {
        setElementAccountSettings(null);
    };

    return (
        <Fragment>
            <Box sx={{ display: "flex", alignItems: "center", textAlign: "center" }}>
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
                        <Avatar sx={{ width: 32, height: 32 }}>
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
                            overflow: "visible",
                            filter: "drop-shadow(0px 2px 8px rgba(0,0,0,0.32))",
                            mt: 1.5,
                            "& .MuiAvatar-root": {
                                width: 32,
                                height: 32,
                                ml: -0.5,
                                mr: 1,
                            },
                            "&:before": {
                                content: '""',
                                display: "block",
                                position: "absolute",
                                top: 0,
                                right: 14,
                                width: 10,
                                height: 10,
                                bgcolor: "background.paper",
                                transform: "translateY(-50%) rotate(45deg)",
                                zIndex: 0,
                            },
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
                    Settings
                </MenuItem>
                <Divider />
                <MenuItem onClick={handleLogoutClick} id={"account-menu-logout"}>
                    <ListItemIcon>
                        <Logout fontSize="small" />
                    </ListItemIcon>
                    Logout
                </MenuItem>
            </Menu>
        </Fragment>
    );
};

export default AccountSettingsMenu;
