import React, { ReactNode, useEffect } from "react";

import SystemSecurityUpdateGoodIcon from "@mui/icons-material/SystemSecurityUpdateGood";
import {
    AppBar,
    Box,
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

import Brand from "@components/Brand";
import { SettingsTwoFactorAuthenticationRoute } from "@constants/Routes";
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

    const navigate = useRouterNavigate();

    return (
        <Box sx={{ display: "flex" }}>
            <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
                <Toolbar variant="dense">
                    <Typography style={{ flexGrow: 1 }}>{translate("Settings")}</Typography>
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
                        <ListItem disablePadding onClick={() => navigate(SettingsTwoFactorAuthenticationRoute)}>
                            <ListItemButton selected={true}>
                                <ListItemIcon>
                                    <SystemSecurityUpdateGoodIcon />
                                </ListItemIcon>
                                <ListItemText primary={translate("Security Keys")} />
                            </ListItemButton>
                        </ListItem>
                    </List>
                </Box>
            </Drawer>
            <Grid container id={props.id} spacing={0}>
                <Grid item xs={12}>
                    <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
                        {props.children}
                    </Box>
                </Grid>
                <Brand />
            </Grid>
        </Box>
    );
};

export default SettingsLayout;
