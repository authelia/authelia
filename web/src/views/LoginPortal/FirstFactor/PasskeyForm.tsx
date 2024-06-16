import React, { Fragment, useCallback, useRef, useState } from "react";

import { Button, CircularProgress, Divider, Grid, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import PasskeyIcon from "@components/PasskeyIcon";
import { RedirectionURL } from "@constants/SearchParams";
import { useIsMountedRef } from "@hooks/Mounted";
import { useQueryParam } from "@hooks/QueryParam";
import { useWorkflow } from "@hooks/Workflow";
import { AssertionResult, AssertionResultFailureString } from "@models/WebAuthn";
import { getWebAuthnPasskeyOptions, getWebAuthnResult, postWebAuthnPasskeyResponse } from "@services/WebAuthn";

export interface Props {
    rememberMe: boolean;

    onAuthenticationStart: () => void;
    onAuthenticationFailure: (err: Error) => void;
    onAuthenticationSuccess: (redirectURL: string | undefined) => void;
}

const PasskeyForm = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const [workflow, workflowID] = useWorkflow();
    const mounted = useIsMountedRef();

    const [loading, setLoading] = useState(false);

    const onSignInErrorCallback = useRef(props.onAuthenticationFailure).current;

    const handleSignIn = useCallback(async () => {
        if (!mounted || loading) {
            return;
        }

        setLoading(true);

        try {
            const optionsStatus = await getWebAuthnPasskeyOptions();

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                setLoading(false);
                onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (result.result !== AssertionResult.Success) {
                if (!mounted.current) return;

                setLoading(false);

                onSignInErrorCallback(new Error(translate(AssertionResultFailureString(result.result))));

                return;
            }

            if (result.response == null) {
                onSignInErrorCallback(
                    new Error(translate("The browser did not respond with the expected attestation data")),
                );
                setLoading(false);

                return;
            }

            if (!mounted.current) return;

            const response = await postWebAuthnPasskeyResponse(result.response, redirectionURL, workflow, workflowID);

            if (response.data.status === "OK" && response.status === 200) {
                props.onAuthenticationSuccess(response.data.data ? response.data.data.redirect : undefined);
                return;
            }

            if (!mounted.current) return;

            onSignInErrorCallback(new Error(translate("The server rejected the security key")));
            setLoading(false);
        } catch (err) {
            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));
            setLoading(false);
        }
    }, [props, mounted, loading, redirectionURL, workflow, workflowID, onSignInErrorCallback, translate]);

    return (
        <Fragment>
            <Grid item xs={12}>
                <Divider component="div" role="presentation">
                    <Typography>OR</Typography>
                </Divider>
            </Grid>
            <Grid item xs={12}>
                <Button
                    id="passkey-sign-in-button"
                    variant="contained"
                    color="primary"
                    fullWidth
                    onClick={handleSignIn}
                    startIcon={<PasskeyIcon />}
                    disabled={loading}
                    endIcon={loading ? <CircularProgress /> : null}
                >
                    {translate("Sign in with a passkey")}
                </Button>
            </Grid>
        </Fragment>
    );
};

export default PasskeyForm;
