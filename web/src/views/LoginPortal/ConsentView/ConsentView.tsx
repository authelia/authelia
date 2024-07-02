import React, { useEffect, useState } from "react";

import { AccountBox, Autorenew, CheckBox, Contacts, Drafts, Group, LockOpen } from "@mui/icons-material";
import {
    Button,
    Checkbox,
    FormControlLabel,
    Grid,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    Theme,
    Tooltip,
    Typography,
} from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";

import { IndexRoute } from "@constants/Routes";
import { Identifier } from "@constants/SearchParams";
import { useNotifications } from "@hooks/NotificationsContext";
import { useRedirector } from "@hooks/Redirector";
import { useUserInfoGET } from "@hooks/UserInfo";
import LoginLayout from "@layouts/LoginLayout";
import { ConsentGetResponseBody, acceptConsent, getConsentResponse, rejectConsent } from "@services/Consent";
import { ComponentOrLoading } from "@views/Generic/ComponentOrLoading";

export interface Props {}

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
        case "authelia.bearer.authz":
            return <LockOpen />;
        default:
            return <CheckBox />;
    }
}

const ConsentView = function (props: Props) {
    const { t: translate } = useTranslation();

    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoGET();

    const { createErrorNotification, resetNotification } = useNotifications();
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();
    const redirect = useRedirector();
    const consentID = searchParams.get(Identifier);

    const [response, setResponse] = useState<ConsentGetResponseBody>();
    const [error, setError] = useState<any>(undefined);
    const [preConfigure, setPreConfigure] = useState(false);

    const styles = useStyles();

    const handlePreConfigureChanged = () => {
        setPreConfigure((preConfigure) => !preConfigure);
    };

    useEffect(() => {
        fetchUserInfo();
    }, [fetchUserInfo]);

    useEffect(() => {
        if (consentID !== null) {
            getConsentResponse(consentID)
                .then((r) => {
                    setResponse(r);
                })
                .catch((error) => {
                    setError(error);
                });
        }
    }, [consentID]);

    useEffect(() => {
        if (error) {
            navigate(IndexRoute);
            console.error(`Unable to display consent screen: ${error.message}`);
        }
    }, [navigate, resetNotification, createErrorNotification, error]);

    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification(translate("There was an issue retrieving user preferences"));
        }
    }, [fetchUserInfoError, resetNotification, createErrorNotification, translate]);

    const translateScopeNameToDescription = (id: string): string => {
        switch (id) {
            case "openid":
                return translate("Use OpenID to verify your identity");
            case "offline_access":
                return translate("Automatically refresh these permissions without user interaction");
            case "profile":
                return translate("Access your profile information");
            case "groups":
                return translate("Access your group membership");
            case "email":
                return translate("Access your email addresses");
            case "authelia.bearer.authz":
                return translate("Access protected resources logged in as you");
            default:
                return id;
        }
    };

    const handleAcceptConsent = async () => {
        // This case should not happen in theory because the buttons are disabled when response is undefined.
        if (!response) {
            return;
        }
        const res = await acceptConsent(preConfigure, response.client_id, consentID);
        if (res.redirect_uri) {
            redirect(res.redirect_uri);
        } else {
            throw new Error("Unable to redirect the user");
        }
    };

    const handleRejectConsent = async () => {
        if (!response) {
            return;
        }
        const res = await rejectConsent(response.client_id, consentID);
        if (res.redirect_uri) {
            redirect(res.redirect_uri);
        } else {
            throw new Error("Unable to redirect the user");
        }
    };

    return (
        <ComponentOrLoading ready={response !== undefined && userInfo !== undefined}>
            <LoginLayout
                id="consent-stage"
                title={`${translate("Hi")} ${userInfo?.display_name}`}
                subtitle={translate("Consent Request")}
            >
                <Grid container>
                    <Grid item xs={12}>
                        <div>
                            <Tooltip
                                title={
                                    translate("Client ID", { client_id: response?.client_id }) ||
                                    "Client ID: " + response?.client_id
                                }
                            >
                                <Typography className={styles.clientDescription}>
                                    {response !== undefined && response.client_description !== ""
                                        ? response.client_description
                                        : response?.client_id}
                                </Typography>
                            </Tooltip>
                        </div>
                    </Grid>
                    <Grid item xs={12}>
                        <div>{translate("The above application is requesting the following permissions")}:</div>
                    </Grid>
                    <Grid item xs={12}>
                        <div className={styles.scopesListContainer}>
                            <List className={styles.scopesList}>
                                {response?.scopes.map((scope: string) => (
                                    <Tooltip title={translate("Scope", { name: scope })}>
                                        <ListItem id={"scope-" + scope} dense>
                                            <ListItemIcon>{scopeNameToAvatar(scope)}</ListItemIcon>
                                            <ListItemText primary={translateScopeNameToDescription(scope)} />
                                        </ListItem>
                                    </Tooltip>
                                ))}
                            </List>
                        </div>
                    </Grid>
                    {response?.pre_configuration ? (
                        <Grid item xs={12}>
                            <Tooltip
                                title={translate("This saves this consent as a pre-configured consent for future use")}
                            >
                                <FormControlLabel
                                    control={
                                        <Checkbox
                                            id="pre-configure"
                                            checked={preConfigure}
                                            onChange={handlePreConfigureChanged}
                                            value="preConfigure"
                                            color="primary"
                                        />
                                    }
                                    className={styles.preConfigure}
                                    label={translate("Remember Consent")}
                                />
                            </Tooltip>
                        </Grid>
                    ) : null}
                    <Grid item xs={12}>
                        <Grid container spacing={1}>
                            <Grid item xs={6}>
                                <Button
                                    id="accept-button"
                                    className={styles.button}
                                    disabled={!response}
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
                                    className={styles.button}
                                    disabled={!response}
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

const useStyles = makeStyles((theme: Theme) => ({
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
    preConfigure: {},
}));

export default ConsentView;
