import React, { useEffect } from "react";

import SystemSecurityUpdateGoodIcon from "@mui/icons-material/SystemSecurityUpdateGood";
import {
    AppBar,
    Box,
    Drawer,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Toolbar,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { IndexRoute } from "@constants/Routes";
import { useAutheliaState } from "@hooks/State";
import { AuthenticationLevel } from "@services/State";

import TwoFactorAuthSettings from "./TwoFactorAuthSettings";

interface Props {}

const drawerWidth = 240;

export default function SettingsView(props: Props) {
    const { t: translate } = useTranslation("settings");
    const navigate = useNavigate();

    const [state, fetchState, , fetchStateError] = useAutheliaState();

    // Fetch the state on page load
    useEffect(() => {
        fetchState();
    }, [fetchState]);

    useEffect(() => {
        if (fetchStateError || (state && state.authentication_level < AuthenticationLevel.OneFactor)) {
            navigate(IndexRoute);
        }
    }, [state, fetchStateError, navigate]);

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
                        <ListItem disablePadding>
                            <ListItemButton selected={true}>
                                <ListItemIcon>
                                    <SystemSecurityUpdateGoodIcon />
                                </ListItemIcon>
                                <ListItemText primary={translate("Two-Factor Authentication")} />
                            </ListItemButton>
                        </ListItem>
                    </List>
                </Box>
            </Drawer>
            <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
                <Toolbar variant="dense" />
                {state && <TwoFactorAuthSettings state={state} />}
            </Box>
        </Box>
    );
}
