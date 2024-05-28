import React, { ReactNode, SyntheticEvent, useCallback, useEffect, useState } from "react";

import { Close, Dashboard, ExpandLess, ExpandMore, Menu, VerifiedUser } from "@mui/icons-material";
//import { Close, Dashboard, Menu, SystemSecurityUpdateGood } from "@mui/icons-material";
import {
    AppBar,
    Box,
    Collapse,
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

import {
    AdminOIDCAuthPoliciesSubRoute,
    AdminOIDCClientSubRoute,
    AdminOIDCProviderSubRoute,
    AdminOIDCSubRoute,
    AdminRoute,
    IndexRoute,
} from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";

export interface Props {
    id?: string;
    children?: ReactNode;
    title?: string;
    titlePrefix?: string;
    drawerWidth?: number;
}

const defaultDrawerWidth = 240;

const AdminLayout = function (props: Props) {
    const { t: translate } = useTranslation("admin");
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
                document.title = `${props.titlePrefix} - ${translate("Admin")} - Authelia`;
            } else {
                document.title = `${translate("Admin")} - Authelia`;
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
                {translate("Admin")}
            </Typography>
            <Divider />
            <List>
                {navItems.map((item) => (
                    <React.Fragment key={item.keyname}>
                        {item.children && item.children !== undefined ? (
                            <DrawerNavItemDropdown
                                key={item.keyname}
                                keyname={item.keyname}
                                text={translate(item.text)}
                                pathname={item.pathname}
                                icon={item.icon}
                                children={item.children}
                            />
                        ) : (
                            <DrawerNavItem
                                key={item.keyname}
                                keyname={item.keyname}
                                text={translate(item.text)}
                                pathname={item.pathname}
                                icon={item.icon}
                            />
                        )}
                    </React.Fragment>
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
                        {translate("Admin")}
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
    children?: NavItem[];
}

const navItems: NavItem[] = [
    { keyname: "overview", text: "Overview", pathname: AdminRoute, icon: <Dashboard color={"primary"} /> },
    {
        keyname: "oidc",
        text: "Open ID Connect",
        pathname: AdminRoute,
        icon: <VerifiedUser color={"primary"} />,
        children: [
            {
                keyname: "oidc-provider",
                text: "Provider",
                pathname: `${AdminRoute}${AdminOIDCSubRoute}${AdminOIDCProviderSubRoute}`,
            },
            {
                keyname: "oidc-clients",
                text: "Clients",
                pathname: `${AdminRoute}${AdminOIDCSubRoute}${AdminOIDCClientSubRoute}`,
            },
            {
                keyname: "oidc-auth-policies",
                text: "Auth. Policies",
                pathname: `${AdminRoute}${AdminOIDCSubRoute}${AdminOIDCAuthPoliciesSubRoute}`,
            },
        ],
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
        <ListItem key={`admin-menu-${props.keyname}`} disablePadding onClick={handleOnClick}>
            <ListItemButton selected={selected} key={`admin-menu-${props.keyname}-button`}>
                {props.icon ? <ListItemIcon>{props.icon}</ListItemIcon> : null}
                <ListItemText primary={props.text} />
            </ListItemButton>
        </ListItem>
    );
};

const DrawerNavItemDropdown = function (props: NavItem) {
    //const selected = window.location.pathname === props.pathname || window.location.pathname === props.pathname + "/";
    const navigate = useRouterNavigate();
    const [open, setOpen] = useState(false);

    const handleOnClick = useCallback(
        (pathname: string) => {
            //console.log("Navigating to:", pathname);
            navigate(pathname);
        },
        [navigate],
    );

    const handleClick = (event: { stopPropagation: () => void }) => {
        event.stopPropagation();
        setOpen(!open);
    };

    const isSelected = (pathname: string) => {
        return window.location.pathname === pathname || window.location.pathname === pathname + "/";
    };

    return (
        <>
            <ListItemButton key={props.keyname} onClick={handleClick}>
                {props.icon && <ListItemIcon>{props.icon}</ListItemIcon>}
                <ListItemText primary={props.text} />
                {open ? <ExpandLess /> : <ExpandMore />}
            </ListItemButton>
            <Collapse in={open} timeout="auto" unmountOnExit>
                <List component="div" disablePadding>
                    {props.children &&
                        props.children.map((child, index) => (
                            <ListItemButton
                                key={`${props.keyname}-${index}`}
                                sx={{ pl: 4, py: 0.3 }}
                                selected={isSelected(child.pathname)}
                                onClick={() => handleOnClick(child.pathname)}
                            >
                                {child.icon && <ListItemIcon>{child.icon}</ListItemIcon>}
                                <ListItemText primary={child.text} sx={{ paddingLeft: 7 }} />
                            </ListItemButton>
                        ))}
                </List>
            </Collapse>
        </>
    );
};

export default AdminLayout;
