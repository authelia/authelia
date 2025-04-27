import React, { Fragment, ReactNode, useCallback, useEffect, useState } from "react";

import { AccountBox, Autorenew, Contacts, Drafts, Group, LockOpen, Policy } from "@mui/icons-material";
import {
    Box,
    Button,
    Checkbox,
    FormControlLabel,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    Theme,
    Tooltip,
    Typography,
    useTheme,
} from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { makeStyles } from "tss-react/mui";

import LogoutButton from "@components/LogoutButton";
import { IndexRoute } from "@constants/Routes";
import { useFlow } from "@hooks/Flow";
import { useNotifications } from "@hooks/NotificationsContext";
import { useRedirector } from "@hooks/Redirector";
import LoginLayout from "@layouts/LoginLayout";
import { UserInfo } from "@models/UserInfo";
import {
    ConsentGetResponseBody,
    acceptConsent,
    formatClaim,
    formatScope,
    getConsentResponse,
    rejectConsent,
} from "@services/ConsentOpenIDConnect";
import { AutheliaState } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";

export interface Props {
    userInfo: UserInfo;
    state: AutheliaState;
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
        case "authelia.bearer.authz":
            return <LockOpen />;
        default:
            return <Policy />;
    }
}

const OpenIDConnectConsentDecisionFormView: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["portal", "consent"]);
    const theme = useTheme();

    const { classes } = useStyles();

    const { createErrorNotification, resetNotification } = useNotifications();
    const navigate = useNavigate();
    const redirect = useRedirector();
    const { id: consentID } = useFlow();

    const [response, setResponse] = useState<ConsentGetResponseBody>();
    const [error, setError] = useState<any>(undefined);
    const [claims, setClaims] = useState<string>("");
    const [preConfigure, setPreConfigure] = useState(false);

    const handlePreConfigureChanged = () => {
        setPreConfigure((preConfigure) => !preConfigure);
    };

    useEffect(() => {
        if (consentID) {
            getConsentResponse(consentID)
                .then((r) => {
                    setResponse(r);
                    setClaims(JSON.stringify(r.claims));
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

    const handleAcceptConsent = async () => {
        // This case should not happen in theory because the buttons are disabled when response is undefined.
        if (!response) {
            return;
        }
        const res = await acceptConsent(preConfigure, response.client_id, JSON.parse(claims), consentID);
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

    const handleClaimCheckboxOnChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setClaims((prevState) => {
            const value = event.target.value;
            const arrClaims: string[] = JSON.parse(prevState);
            const checking = !arrClaims.includes(event.target.value);

            if (checking) {
                if (!arrClaims.includes(value)) {
                    arrClaims.push(value);
                }
            } else {
                const i = arrClaims.indexOf(value);

                if (i > -1) {
                    arrClaims.splice(i, 1);
                }
            }

            return JSON.stringify(arrClaims);
        });
    };

    const claimChecked = useCallback(
        (claim: string) => {
            const arrClaims: string[] = JSON.parse(claims);

            return arrClaims.includes(claim);
        },
        [claims],
    );

    const hasClaims = response?.essential_claims || response?.claims;

    return (
        <ComponentOrLoading ready={response !== undefined}>
            <LoginLayout
                id="consent-stage"
                title={`${translate("Hi")} ${props.userInfo.display_name}`}
                subtitle={translate("Consent Request")}
            >
                <Grid container direction={"column"} justifyContent={"center"} alignItems={"center"}>
                    <Grid size={{ xs: 12 }} sx={{ paddingBottom: theme.spacing(2) }}>
                        <LogoutButton />
                    </Grid>
                    <Grid size={{ xs: 12 }}>
                        <Grid container alignItems={"center"} justifyContent="center">
                            <Grid size={{ xs: 12 }}>
                                <Box>
                                    <Tooltip
                                        title={
                                            translate("Client ID", { client_id: response?.client_id }) ||
                                            "Client ID: " + response?.client_id
                                        }
                                    >
                                        <Typography className={classes.clientDescription}>
                                            {response !== undefined && response.client_description !== ""
                                                ? response.client_description
                                                : response?.client_id}
                                        </Typography>
                                    </Tooltip>
                                </Box>
                            </Grid>
                            <Grid size={{ xs: 12 }}>
                                <Box>{translate("The above application is requesting the following permissions")}:</Box>
                            </Grid>
                            <Grid size={{ xs: 12 }}>
                                <Box className={classes.scopesListContainer}>
                                    <List className={classes.scopesList}>
                                        {response?.scopes.map((scope: string) => (
                                            <Tooltip title={translate("Scope", { name: scope, ns: "consent" })}>
                                                <ListItem id={"scope-" + scope} dense>
                                                    <ListItemIcon>{scopeNameToAvatar(scope)}</ListItemIcon>
                                                    <ListItemText
                                                        primary={formatScope(
                                                            translate(`scopes.${scope}`, { ns: "consent" }),
                                                            scope,
                                                        )}
                                                    />
                                                </ListItem>
                                            </Tooltip>
                                        ))}
                                    </List>
                                </Box>
                            </Grid>
                            {hasClaims ? (
                                <Grid size={{ xs: 12 }}>
                                    <Box className={classes.claimsListContainer}>
                                        <List className={classes.claimsList}>
                                            {response?.essential_claims?.map((claim: string) => (
                                                <Tooltip title={translate("Claim", { name: claim, ns: "consent" })}>
                                                    <FormControlLabel
                                                        control={
                                                            <Checkbox
                                                                id={`claim-${claim}-essential`}
                                                                disabled
                                                                checked
                                                            />
                                                        }
                                                        label={formatClaim(
                                                            translate(`claims.${claim}`, { ns: "consent" }),
                                                            claim,
                                                        )}
                                                    />
                                                </Tooltip>
                                            ))}
                                            {response?.claims?.map((claim: string) => (
                                                <Tooltip title={translate("Claim", { name: claim, ns: "consent" })}>
                                                    <FormControlLabel
                                                        control={
                                                            <Checkbox
                                                                id={"claim-" + claim}
                                                                value={claim}
                                                                checked={claimChecked(claim)}
                                                                onChange={handleClaimCheckboxOnChange}
                                                            />
                                                        }
                                                        label={formatClaim(
                                                            translate(`claims.${claim}`, { ns: "consent" }),
                                                            claim,
                                                        )}
                                                    />
                                                </Tooltip>
                                            ))}
                                        </List>
                                    </Box>
                                </Grid>
                            ) : null}
                            {response?.pre_configuration ? (
                                <Grid size={{ xs: 12 }}>
                                    <Tooltip
                                        title={translate(
                                            "This saves this consent as a pre-configured consent for future use",
                                        )}
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
                                            className={classes.preConfigure}
                                            label={translate("Remember Consent")}
                                        />
                                    </Tooltip>
                                </Grid>
                            ) : null}
                            <Grid size={{ xs: 12 }}>
                                <Grid container spacing={1}>
                                    <Grid size={{ xs: 6 }}>
                                        <Button
                                            id="accept-button"
                                            className={classes.button}
                                            disabled={!response}
                                            onClick={handleAcceptConsent}
                                            color="primary"
                                            variant="contained"
                                        >
                                            {translate("Accept")}
                                        </Button>
                                    </Grid>
                                    <Grid size={{ xs: 6 }}>
                                        <Button
                                            id="deny-button"
                                            className={classes.button}
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
                    </Grid>
                </Grid>
            </LoginLayout>
        </ComponentOrLoading>
    );
};

interface ComponentOrLoadingProps {
    ready: boolean;

    children: ReactNode;
}

function ComponentOrLoading(props: ComponentOrLoadingProps) {
    return (
        <Fragment>
            <Box className={props.ready ? "hidden" : ""}>
                <LoadingPage />
            </Box>
            {props.ready ? props.children : null}
        </Fragment>
    );
}

const useStyles = makeStyles()((theme: Theme) => ({
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
    claimsListContainer: {
        textAlign: "center",
    },
    claimsList: {
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

export default OpenIDConnectConsentDecisionFormView;
