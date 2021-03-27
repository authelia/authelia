import React, { useEffect } from "react";

import { Grid, Button, List, ListItem, ListItemText, ListItemIcon, makeStyles } from "@material-ui/core";
import { AccountBox, CheckBox, Contacts, Drafts, Group } from "@material-ui/icons";
import { useHistory } from "react-router-dom";

import { useRequestedScopes } from "../../../hooks/Consent";
import { useNotifications } from "../../../hooks/NotificationsContext";
import { useRedirector } from "../../../hooks/Redirector";
import LoginLayout from "../../../layouts/LoginLayout";
import { acceptConsent, rejectConsent } from "../../../services/Consent";

export interface Props {}

function showListItemAvatar(id: string) {
    switch (id) {
        case "openid":
            return <AccountBox />;
        case "profile":
            return <Contacts />;
        case "groups":
            return <Group />;
        case "email":
            return <Drafts />;
        default:
            return <CheckBox />;
    }
}

const ConsentView = function (props: Props) {
    const classes = useStyles();
    const history = useHistory();
    const redirect = useRedirector();
    const { createErrorNotification, resetNotification } = useNotifications();
    const [resp, fetch, , err] = useRequestedScopes();

    useEffect(() => {
        if (err) {
            createErrorNotification(err.message);

            // If there is an error we simply redirect to the main login page.
            setTimeout(() => {
                resetNotification();
                history.push("/");
            }, 1000);
        }
    }, [history, resetNotification, createErrorNotification, err]);

    useEffect(() => {
        fetch();
    }, [fetch]);

    const handleAcceptConsent = async () => {
        // This case should not happen in theory because the buttons are disabled when response is undefined.
        if (!resp) {
            return;
        }
        const res = await acceptConsent(resp.client_id);
        if (res.redirect_uri) {
            redirect(res.redirect_uri);
        } else {
            throw new Error("Unable to redirect the user");
        }
    };

    const handleRejectConsent = async () => {
        if (!resp) {
            return;
        }
        const res = await rejectConsent(resp.client_id);
        if (res.redirect_uri) {
            redirect(res.redirect_uri);
        } else {
            throw new Error("Unable to redirect the user");
        }
    };

    return (
        <LoginLayout id="consent-stage" title={`Permissions Request`} showBrand>
            <Grid container>
                <Grid item xs={12}>
                    <div>
                        The application {resp?.client_description} ({resp?.client_id}) is requesting the following
                        permissions:
                    </div>
                </Grid>
                <Grid item xs={12}>
                    <List className={classes.scopesList}>
                        {resp?.scopes.map((s) => (
                            <ListItem id={s.name}>
                                <ListItemIcon>{showListItemAvatar(s.name)}</ListItemIcon>
                                <ListItemText primary={s.description} />
                            </ListItem>
                        ))}
                    </List>
                </Grid>
                <Grid item xs={12}>
                    <div>
                        <Button
                            className={classes.button}
                            disabled={!resp}
                            onClick={handleAcceptConsent}
                            color="primary"
                            variant="contained"
                        >
                            Accept
                        </Button>
                        <Button
                            className={classes.button}
                            disabled={!resp}
                            onClick={handleRejectConsent}
                            color="secondary"
                            variant="contained"
                        >
                            Deny
                        </Button>
                    </div>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

const useStyles = makeStyles((theme) => ({
    container: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
        display: "block",
        justifyContent: "center",
    },
    scopesList: {
        backgroundColor: theme.palette.background.paper,
    },
    clientID: {
        fontWeight: "bold",
    },
    button: {
        marginLeft: theme.spacing(),
        marginRight: theme.spacing(),
    },
    bulletIcon: {
        display: "inline-block",
    },
    permissionsContainer: {
        border: "1px solid #dedede",
        margin: theme.spacing(4),
    },
    listItem: {
        textAlign: "left",
        marginRight: theme.spacing(2),
    },
}));

export default ConsentView;
