import React from "react";

import {
    AccountBox,
    Autorenew,
    Contacts,
    Drafts,
    Group,
    Home,
    LockOpen,
    PhoneAndroid,
    Policy,
} from "@mui/icons-material";
import { Box, List, ListItem, ListItemIcon, ListItemText, Theme, Tooltip } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import { formatScope } from "@services/ConsentOpenIDConnect";

export interface Props {
    scopes: string[];
}

function scopeNameToAvatar(id: string) {
    switch (id) {
        case "openid":
            return <AccountBox />;
        case "offline_access":
            return <Autorenew />;
        case "profile":
            return <Contacts />;
        case "groups":
            return <Group />;
        case "email":
            return <Drafts />;
        case "phone":
            return <PhoneAndroid />;
        case "address":
            return <Home />;
        case "authelia.bearer.authz":
            return <LockOpen />;
        default:
            return <Policy />;
    }
}

const DecisionFormScopes: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent"]);

    const { classes } = useStyles();

    return (
        <Grid size={{ xs: 12 }}>
            <Box className={classes.scopesListContainer}>
                <List className={classes.scopesList}>
                    {props.scopes.map((scope: string) => (
                        <Tooltip title={translate("Scope", { name: scope })}>
                            <ListItem id={"scope-" + scope} dense>
                                <ListItemIcon>{scopeNameToAvatar(scope)}</ListItemIcon>
                                <ListItemText primary={formatScope(translate(`scopes.${scope}`), scope)} />
                            </ListItem>
                        </Tooltip>
                    ))}
                </List>
            </Box>
        </Grid>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    scopesListContainer: {
        textAlign: "center",
    },
    scopesList: {
        display: "inline-block",
        backgroundColor: theme.palette.background.paper,
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
}));

export default DecisionFormScopes;
