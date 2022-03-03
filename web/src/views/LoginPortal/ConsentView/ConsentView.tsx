import React, { useEffect, Fragment, ReactNode } from "react";

import {
    Button,
    Grid,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    Tooltip,
    Typography,
    makeStyles,
} from "@material-ui/core";
import { AccountBox, CheckBox, Contacts, Drafts, Group } from "@material-ui/icons";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { IndexRoute } from "@constants/Routes";
import { useRequestedScopes } from "@hooks/Consent";
import { useNotifications } from "@hooks/NotificationsContext";
import { useRedirector } from "@hooks/Redirector";
import LoginLayout from "@layouts/LoginLayout";
import { acceptConsent, rejectConsent } from "@services/Consent";
import LoadingPage from "@views/LoadingPage/LoadingPage";

export interface Props {}

function scopeNameToAvatar(id: string) {
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
    const { t: translate } = useTranslation("Portal");

    useEffect(() => {
        if (err) {
            navigate(IndexRoute);
            console.error(`Unable to display consent screen: ${err.message}`);
        }
    }, [navigate, resetNotification, createErrorNotification, err]);

    useEffect(() => {
        fetch();
    }, [fetch]);

    const translateScopeNameToDescription = (id: string): string => {
        switch (id) {
            case "openid":
                return translate("Use OpenID to verify your identity");
            case "profile":
                return translate("Access your display name");
            case "groups":
                return translate("Access your group membership");
            case "email":
                return translate("Access your email addresses");
            default:
                return id;
        }
    };

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
                        <div>
                            {resp !== undefined && resp.client_description !== "" ? (
                                <Tooltip title={"Client ID: " + resp.client_id}>
                                    <Typography className={classes.clientDescription}>
                                        {resp.client_description}
                                    </Typography>
                                </Tooltip>
                            ) : (
                                <Tooltip title={"Client ID: " + resp?.client_id}>
                                    <Typography className={classes.clientDescription}>{resp?.client_id}</Typography>
                                </Tooltip>
                            )}
                        </div>
                    </Grid>
                    <Grid item xs={12}>
                        <div>{translate("The above application is requesting the following permissions")}:</div>
                    </Grid>
                    <Grid item xs={12}>
                        <div className={classes.scopesListContainer}>
                            <List className={classes.scopesList}>
                                {resp?.scopes.map((scope: string) => (
                                    <Tooltip title={"Scope " + scope}>
                                        <ListItem id={"scope-" + scope} dense>
                                            <ListItemIcon>{scopeNameToAvatar(scope)}</ListItemIcon>
                                            <ListItemText primary={translateScopeNameToDescription(scope)} />
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
                                    {translate("Accept")}
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
                                    {translate("Deny")}
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
    clientDescription: {
        fontWeight: 600,
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
