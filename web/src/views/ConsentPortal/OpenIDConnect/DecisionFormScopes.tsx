import { FC } from "react";

import { Box, List, ListItem, ListItemIcon, ListItemText, Tooltip } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import { ScopeAvatar } from "@components/OpenIDConnect";
import { formatScope } from "@services/ConsentOpenIDConnect";

export interface Props {
    scopes: string[];
}

const DecisionFormScopes: FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent"]);

    return (
        <Grid size={{ xs: 12 }}>
            <Box sx={{ textAlign: "center" }}>
                <List
                    sx={{
                        backgroundColor: (theme) => theme.palette.background.paper,
                        display: "inline-block",
                        marginBottom: (theme) => theme.spacing(2),
                        marginTop: (theme) => theme.spacing(2),
                    }}
                >
                    {props.scopes.map((scope: string) => (
                        <Tooltip key={scope} title={translate("Scope", { name: scope })}>
                            <ListItem id={"scope-" + scope} key={scope} dense>
                                <ListItemIcon>{ScopeAvatar(scope)}</ListItemIcon>
                                <ListItemText primary={formatScope(translate(`scopes.${scope}`), scope)} />
                            </ListItem>
                        </Tooltip>
                    ))}
                </List>
            </Box>
        </Grid>
    );
};

export default DecisionFormScopes;
