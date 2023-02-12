import React, { ReactNode, useEffect } from "react";

import { Dashboard } from "@mui/icons-material";
import SystemSecurityUpdateGoodIcon from "@mui/icons-material/SystemSecurityUpdateGood";
import {
    AppBar,
    Box,
    Button,
    Drawer,
    Grid,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Toolbar,
    Typography,
} from "@mui/material";
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

    const navigate = useRouterNavigate();

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

    return (
        <Box sx={{ display: "flex" }}>
            <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
                <Toolbar variant="dense">
                    <Typography style={{ flexGrow: 1 }}>Authelia {translate("Settings")}</Typography>
                    <Button
                        variant="contained"
                        color="success"
                        onClick={() => {
                            navigate(IndexRoute);
                        }}
                    >
                        {translate("Close")}
                    </Button>
                </Toolbar>
            </AppBar>
            <Drawer
                variant="permanent"
                sx={{
                    width: drawerWidth,
                    flexShrink: 0,
                    [`& .MuiDrawer-paper`]: { width: drawerWidth, boxSizing: "border-box" },
                }}
            >
                <Toolbar variant="dense" />
                <Box sx={{ overflow: "auto" }}>
                    <List>
                        <SettingsMenuItem pathname={SettingsRoute} text={translate("Overview")} icon={<Dashboard />} />
                        <SettingsMenuItem
                            pathname={`${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`}
                            text={translate("Two-Factor Authentication")}
                            icon={<SystemSecurityUpdateGoodIcon />}
                        />
                    </List>
                </Box>
            </Drawer>
            <Grid container id={props.id} spacing={0}>
                <Grid item xs={12}>
                    <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
                        <Toolbar variant="dense" />
                        {props.children}
                    </Box>
                </Grid>
            </Grid>
        </Box>
    );
};

export default SettingsLayout;

interface SettingsMenuItemProps {
    pathname: string;
    text: string;
    icon: ReactNode;
}

const SettingsMenuItem = function (props: SettingsMenuItemProps) {
    const selected = window.location.pathname === props.pathname || window.location.pathname === props.pathname + "/";
    const navigate = useRouterNavigate();

    return (
        <ListItem disablePadding onClick={!selected ? () => navigate(props.pathname) : undefined}>
            <ListItemButton selected={selected}>
                <ListItemIcon>{props.icon}</ListItemIcon>
                <ListItemText primary={props.text} />
            </ListItemButton>
        </ListItem>
    );
};
