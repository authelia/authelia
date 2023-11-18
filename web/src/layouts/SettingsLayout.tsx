import React, { ReactNode, SyntheticEvent, useCallback, useEffect, useState } from "react";

import { Close, Dashboard, Menu, SystemSecurityUpdateGood } from "@mui/icons-material";
import {
    AppBar,
    Box,
    Divider,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    SwipeableDrawer,
    Toolbar,
    Typography,
} from "@mui/material";
import IconButton from "@mui/material/IconButton";
import { useTranslation } from "react-i18next";

import { IndexRoute, SettingsRoute, SettingsTwoFactorAuthenticationSubRoute } from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: string;
    titlePrefix?: string;
    drawerWidth?: number;
}

const defaultDrawerWidth = 240;

const SettingsLayout = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const [drawerOpen, setDrawerOpen] = useState(false);

    useEffect(() => {
        if (props.title) {
            if (props.titlePrefix) {
                document.title = `${props.titlePrefix} - ${props.title} - Authelia`;
            } else {
                document.title = `${props.title} - Authelia`;
            }
        } else {
            if (props.titlePrefix) {
                document.title = `${props.titlePrefix} - ${translate("Settings")} - Authelia`;
            } else {
                document.title = `${translate("Settings")} - Authelia`;
            }
        }
    }, [props.title, props.titlePrefix, translate]);

    const drawerWidth = props.drawerWidth === undefined ? defaultDrawerWidth : props.drawerWidth;

    const handleToggleDrawer = (event: SyntheticEvent) => {
        if (
            event.nativeEvent instanceof KeyboardEvent &&
            event.nativeEvent.type === "keydown" &&
            (event.nativeEvent.key === "Tab" || event.nativeEvent.key === "Shift")
        ) {
            return;
        }

        setDrawerOpen((state) => !state);
    };

    const container = window !== undefined ? () => window.document.body : undefined;

    const drawer = (
        <Box onClick={handleToggleDrawer} sx={{ textAlign: "center" }}>
            <Typography variant="h6" sx={{ my: 2 }}>
                {translate("Settings")}
            </Typography>
            <Divider />
            <List>
                {navItems.map((item) => (
                    <DrawerNavItem
                        key={item.keyname}
                        keyname={item.keyname}
                        text={translate(item.text)}
                        pathname={item.pathname}
                        icon={item.icon}
                    />
                ))}
            </List>
        </Box>
    );

    return (
        <Box sx={{ display: "flex" }}>
            <AppBar component={"nav"}>
                <Toolbar>
                    <IconButton
                        id={"settings-menu"}
                        edge={"start"}
                        color={"inherit"}
                        aria-label={"open drawer"}
                        onClick={handleToggleDrawer}
                        sx={{ mr: 2 }}
                    >
                        <Menu />
                    </IconButton>
                    <Typography
                        variant={"h6"}
                        component={"div"}
                        sx={{ flexGrow: 1, display: { xs: drawerOpen ? "none" : "block" } }}
                    >
                        {translate("Settings")}
                    </Typography>
                </Toolbar>
            </AppBar>
            <Box component={"nav"}>
                <SwipeableDrawer
                    container={container}
                    anchor={"left"}
                    open={drawerOpen}
                    onOpen={handleToggleDrawer}
                    onClose={handleToggleDrawer}
                    ModalProps={{
                        keepMounted: true,
                    }}
                    sx={{
                        display: { xs: "block" },
                        "& .MuiDrawer-paper": { boxSizing: "border-box", width: drawerWidth },
                    }}
                >
                    {drawer}
                </SwipeableDrawer>
            </Box>
            <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
                <Toolbar />
                {props.children}
            </Box>
        </Box>
    );
};

interface NavItem {
    keyname: string;
    text: string;
    pathname: string;
    icon?: ReactNode;
}

const navItems: NavItem[] = [
    { keyname: "overview", text: "Overview", pathname: SettingsRoute, icon: <Dashboard color={"primary"} /> },
    {
        keyname: "twofactor",
        text: "Two-Factor Authentication",
        pathname: `${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`,
        icon: <SystemSecurityUpdateGood color={"primary"} />,
    },
    { keyname: "close", text: "Close", pathname: IndexRoute, icon: <Close color={"error"} /> },
];

const DrawerNavItem = function (props: NavItem) {
    const selected = window.location.pathname === props.pathname || window.location.pathname === props.pathname + "/";
    const navigate = useRouterNavigate();

    const handleOnClick = useCallback(() => {
        if (selected) {
            return;
        }

        navigate(props.pathname);
    }, [navigate, props, selected]);

    return (
        <ListItem disablePadding onClick={handleOnClick}>
            <ListItemButton selected={selected} id={`settings-menu-${props.keyname}`}>
                {props.icon ? <ListItemIcon>{props.icon}</ListItemIcon> : null}
                <ListItemText primary={props.text} />
            </ListItemButton>
        </ListItem>
    );
};

export default SettingsLayout;
