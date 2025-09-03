import React from "react";

import { Box, List, ListItem, ListItemIcon, ListItemText, Theme, Tooltip } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import { ScopeAvatar } from "@components/OpenIDConnect";
import { formatScope } from "@services/ConsentOpenIDConnect";

export interface Props {
    scopes: string[];
}

const DecisionFormScopes: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent"]);

    const { classes } = useStyles();

    return (
        <Grid size={{ xs: 12 }}>
            <Box className={classes.scopesListContainer}>
                <List className={classes.scopesList}>
                    {props.scopes.map((scope: string) => (
                        <Tooltip key={scope} title={translate("Scope", { name: scope })}>
                            <ListItem id={"scope-" + scope} key={scope} dense>
                                <ListItemIcon>{ScopeAvatar(scope)}</ListItemIcon>
                                <ListItemText primary={formatScope(scope)} />
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
