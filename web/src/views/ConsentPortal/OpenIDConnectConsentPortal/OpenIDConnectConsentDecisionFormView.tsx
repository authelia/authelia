import React, { Fragment, useEffect, useState } from "react";

import { Box, Button, Theme, Tooltip, Typography, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import LogoutButton from "@components/LogoutButton";
import { IndexRoute } from "@constants/Routes";
import { useFlow } from "@hooks/Flow";
import { useNotifications } from "@hooks/NotificationsContext";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useRedirector } from "@hooks/Redirector";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import LoginLayout from "@layouts/LoginLayout";
import { UserInfo } from "@models/UserInfo";
import {
    ConsentGetResponseBody,
    getConsentResponse,
    postConsentResponseAccept,
    postConsentResponseReject,
} from "@services/ConsentOpenIDConnect";
import { AutheliaState, AuthenticationLevel } from "@services/State";
import OpenIDConnectConsentDecisionFormClaims from "@views/ConsentPortal/OpenIDConnectConsentPortal/OpenIDConnectConsentDecisionFormClaims.js";
import OpenIDConnectConsentDecisionFormPreConfiguration from "@views/ConsentPortal/OpenIDConnectConsentPortal/OpenIDConnectConsentDecisionFormPreConfiguration.js";
import OpenIDConnectConsentDecisionFormScopes from "@views/ConsentPortal/OpenIDConnectConsentPortal/OpenIDConnectConsentDecisionFormScopes.js";
import LoadingPage from "@views/LoadingPage/LoadingPage";

export interface Props {
    userInfo?: UserInfo;
    state: AutheliaState;
}

const OpenIDConnectConsentDecisionFormView: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["portal", "consent"]);
    const theme = useTheme();

    const { classes } = useStyles();

    const { createErrorNotification, resetNotification } = useNotifications();
    const navigate = useRouterNavigate();
    const redirect = useRedirector();
    const { id: flowID, subflow } = useFlow();
    const userCode = useUserCode();

    const [response, setResponse] = useState<ConsentGetResponseBody>();
    const [error, setError] = useState<any>(undefined);
    const [claims, setClaims] = useState<string[]>([]);
    const [preConfigure, setPreConfigure] = useState(false);

    const handlePreConfigureChanged = (value: boolean) => {
        setPreConfigure(value);
    };

    useEffect(() => {
        if (props.state.authentication_level === AuthenticationLevel.Unauthenticated) {
            navigate(IndexRoute);
        } else if (flowID || userCode) {
            getConsentResponse(flowID, userCode)
                .then((r) => {
                    setResponse(r);
                })
                .catch((error) => {
                    setError(error);
                });
        } else {
            navigate(IndexRoute);
        }
    }, [flowID, navigate, props.state.authentication_level, userCode]);

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

        const res = await postConsentResponseAccept(
            preConfigure,
            response.client_id,
            claims,
            flowID,
            subflow,
            userCode,
        );

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
        const res = await postConsentResponseReject(response.client_id, flowID, subflow, userCode);
        if (res.redirect_uri) {
            redirect(res.redirect_uri);
        } else {
            throw new Error("Unable to redirect the user");
        }
    };

    return (
        <Fragment>
            {props.userInfo && response !== undefined ? (
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
                                                {response.client_description !== ""
                                                    ? response.client_description
                                                    : response?.client_id}
                                            </Typography>
                                        </Tooltip>
                                    </Box>
                                </Grid>
                                <Grid size={{ xs: 12 }}>
                                    <Box>
                                        {translate("The above application is requesting the following permissions")}:
                                    </Box>
                                </Grid>
                                <OpenIDConnectConsentDecisionFormScopes scopes={response.scopes} />
                                <OpenIDConnectConsentDecisionFormClaims
                                    claims={response.claims}
                                    essential_claims={response.essential_claims}
                                    onChangeChecked={(claims) => setClaims(claims)}
                                />
                                <OpenIDConnectConsentDecisionFormPreConfiguration
                                    pre_configuration={response.pre_configuration}
                                    onChangePreConfiguration={handlePreConfigureChanged}
                                />
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
            ) : (
                <Box>
                    <LoadingPage />
                </Box>
            )}
        </Fragment>
    );
};

const useStyles = makeStyles()((theme: Theme) => ({
    clientDescription: {
        fontWeight: 600,
    },
    button: {
        marginLeft: theme.spacing(),
        marginRight: theme.spacing(),
        width: "100%",
    },
}));

export default OpenIDConnectConsentDecisionFormView;
