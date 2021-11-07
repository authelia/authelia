import React, { useEffect, Fragment, ReactNode } from "react";

import { Button, Grid, List, ListItem, ListItemIcon, ListItemText, Tooltip, makeStyles } from "@material-ui/core";
import { AccountBox, CheckBox, Contacts, Drafts, Group } from "@material-ui/icons";
import { useNavigate } from "react-router-dom";

import { FirstFactorRoute } from "@constants/Routes";
import { useRequestedScopes } from "@hooks/Consent";
import { useNotifications } from "@hooks/NotificationsContext";
import { useRedirector } from "@hooks/Redirector";
import LoginLayout from "@layouts/LoginLayout";
import { acceptConsent, rejectConsent } from "@services/Consent";
import LoadingPage from "@views/LoadingPage/LoadingPage";

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
    const navigate = useNavigate();
    const redirect = useRedirector();
    const { createErrorNotification, resetNotification } = useNotifications();
    const [resp, fetch, , err] = useRequestedScopes();

    useEffect(() => {
        if (err) {
            navigate(FirstFactorRoute);
            console.error(`Unable to display consent screen: ${err.message}`);
        }
    }, [navigate, resetNotification, createErrorNotification, err]);

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
        <ComponentOrLoading ready={resp !== undefined}>
            <LoginLayout id="consent-stage" title={`Permissions Request`} showBrand>
                <Grid container>
                    <Grid item xs={12}>
                        <div style={{ textAlign: "left" }}>
                            The application
                            <b>{` ${resp?.client_description} (${resp?.client_id}) `}</b>
                            is requesting the following permissions
                        </div>
                    </Grid>
                    <Grid item xs={12}>
                        <div className={classes.scopesListContainer}>
                            <List className={classes.scopesList}>
                                {resp?.scopes.map((s) => (
                                    <Tooltip title={"Scope " + s.name}>
                                        <ListItem id={"scope-" + s.name} dense>
                                            <ListItemIcon>{showListItemAvatar(s.name)}</ListItemIcon>
                                            <ListItemText primary={s.description} />
                                        </ListItem>
                                    </Tooltip>
                                ))}
                            </List>
                        </div>
                    </Grid>
                    <Grid item xs={12}>
                        <Grid container spacing={1}>
                            <Grid item xs={6}>
                                <Button
                                    id="accept-button"
                                    className={classes.button}
                                    disabled={!resp}
                                    onClick={handleAcceptConsent}
                                    color="primary"
                                    variant="contained"
                                >
                                    Accept
                                </Button>
                            </Grid>
                            <Grid item xs={6}>
                                <Button
                                    id="deny-button"
                                    className={classes.button}
                                    disabled={!resp}
                                    onClick={handleRejectConsent}
                                    color="secondary"
                                    variant="contained"
                                >
                                    Deny
                                </Button>
                            </Grid>
                        </Grid>
                    </Grid>
                </Grid>
            </LoginLayout>
        </ComponentOrLoading>
    );
};

const useStyles = makeStyles((theme) => ({
    container: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
        display: "block",
        justifyContent: "center",
    },
    scopesListContainer: {
        textAlign: "center",
    },
    scopesList: {
        display: "inline-block",
        backgroundColor: theme.palette.background.paper,
        marginTop: theme.spacing(2),
        marginBottom: theme.spacing(2),
    },
    clientID: {
        fontWeight: "bold",
    },
    button: {
        marginLeft: theme.spacing(),
        marginRight: theme.spacing(),
        width: "100%",
    },
    bulletIcon: {
        display: "inline-block",
    },
    permissionsContainer: {
        border: "1px solid #dedede",
        margin: theme.spacing(4),
    },
    listItem: {
        textAlign: "center",
        marginRight: theme.spacing(2),
    },
}));

export default ConsentView;

interface ComponentOrLoadingProps {
    ready: boolean;

    children: ReactNode;
}

function ComponentOrLoading(props: ComponentOrLoadingProps) {
    return (
        <Fragment>
            <div className={props.ready ? "hidden" : ""}>
                <LoadingPage />
            </div>
            {props.ready ? props.children : null}
        </Fragment>
    );
}
