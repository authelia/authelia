import { ReactNode, SyntheticEvent, useCallback, useEffect, useState } from "react";

import { Close, Dashboard, Menu, Security, SystemSecurityUpdateGood } from "@mui/icons-material";
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

import { EncodedName } from "@constants/constants";
import {
    IndexRoute,
    SecuritySubRoute,
    SettingsRoute,
    SettingsTwoFactorAuthenticationSubRoute,
} from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";

export interface Props {
    children?: ReactNode;
    drawerWidth?: number;
}

const defaultDrawerWidth = 240;

const SettingsLayout = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const [drawerOpen, setDrawerOpen] = useState(false);

    useEffect(() => {
        document.title = translate("Settings - {{authelia}}", { authelia: atob(String.fromCodePoint(...EncodedName)) });
    }, [translate]);

    const drawerWidth = props.drawerWidth ?? defaultDrawerWidth;

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

    const container = typeof globalThis === "undefined" ? undefined : () => globalThis.document.body;

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
                        sx={{ display: { xs: drawerOpen ? "none" : "block" }, flexGrow: 1 }}
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
                        "& .MuiDrawer-paper": { boxSizing: "border-box", width: drawerWidth },
                        display: { xs: "block" },
                    }}
                >
                    {drawer}
                </SwipeableDrawer>
            </Box>
            <Box component="main" sx={{ flexGrow: 1, p: { sm: 3, xs: 0 } }}>
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
    { icon: <Dashboard color={"primary"} />, keyname: "overview", pathname: SettingsRoute, text: "Overview" },
    {
        icon: <Security color={"primary"} />,
        keyname: "security",
        pathname: `${SettingsRoute}${SecuritySubRoute}`,
        text: "Security",
    },
    {
        icon: <SystemSecurityUpdateGood color={"primary"} />,
        keyname: "twofactor",
        pathname: `${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`,
        text: "Two-Factor Authentication",
    },
    { icon: <Close color={"error"} />, keyname: "close", pathname: IndexRoute, text: "Close" },
];

const DrawerNavItem = function (props: NavItem) {
    const selected =
        globalThis.location.pathname === props.pathname || globalThis.location.pathname === props.pathname + "/";
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
