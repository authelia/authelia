import React, { Fragment, useCallback, useRef, useState } from "react";

import { Button, CircularProgress, Divider, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import PasskeyIcon from "@components/PasskeyIcon";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useIsMountedRef } from "@hooks/Mounted";
import { useQueryParam } from "@hooks/QueryParam";
import { AssertionResult, AssertionResultFailureString } from "@models/WebAuthn";
import { getWebAuthnPasskeyOptions, getWebAuthnResult, postWebAuthnPasskeyResponse } from "@services/WebAuthn";

export interface Props {
    disabled: boolean;
    rememberMe: boolean;

    onAuthenticationStart: () => void;
    onAuthenticationStop: () => void;
    onAuthenticationError: (err: Error) => void;
    onAuthenticationSuccess: (redirectURL: string | undefined) => void;
}

const PasskeyForm = function (props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const requestMethod = useQueryParam(RequestMethod);
    const { id: flowID, flow, subflow } = useFlow();
    const mounted = useIsMountedRef();

    const [loading, setLoading] = useState(false);

    const onSignInErrorCallback = useRef(props.onAuthenticationError).current;

    const handleAuthenticationStart = useCallback(() => {
        props.onAuthenticationStart();
        setLoading(true);
    }, [props]);

    const handleAuthenticationStop = useCallback(() => {
        props.onAuthenticationStop();
        setLoading(false);
    }, [props]);

    const handleSignIn = useCallback(async () => {
        if (!mounted || loading) {
            return;
        }

        handleAuthenticationStart();

        try {
            const optionsStatus = await getWebAuthnPasskeyOptions();

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                handleAuthenticationStop();
                onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));

                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (result.result !== AssertionResult.Success) {
                if (!mounted.current) return;

                handleAuthenticationStop();

                onSignInErrorCallback(new Error(translate(AssertionResultFailureString(result.result))));

                return;
            }

            if (result.response == null) {
                onSignInErrorCallback(
                    new Error(translate("The browser did not respond with the expected attestation data")),
                );
                handleAuthenticationStop();

                return;
            }

            if (!mounted.current) return;

            const response = await postWebAuthnPasskeyResponse(
                result.response,
                props.rememberMe,
                redirectionURL,
                requestMethod,
                flowID,
                flow,
                subflow,
            );

            handleAuthenticationStop();

            if (response.data.status === "OK" && response.status === 200) {
                props.onAuthenticationSuccess(response.data.data ? response.data.data.redirect : undefined);
                return;
            }

            if (!mounted.current) return;

            onSignInErrorCallback(new Error(translate("The server rejected the security key")));
        } catch (err) {
            handleAuthenticationStop();

            // If the request was initiated and the user changed 2FA method in the meantime,
            // the process is interrupted to avoid updating state of unmounted component.
            if (!mounted.current) return;
            console.error(err);
            onSignInErrorCallback(new Error(translate("Failed to initiate security key sign in process")));
        }
    }, [
        mounted,
        loading,
        handleAuthenticationStart,
        props,
        redirectionURL,
        requestMethod,
        flowID,
        flow,
        subflow,
        handleAuthenticationStop,
        onSignInErrorCallback,
        translate,
    ]);

    return (
        <Fragment>
            <Grid size={{ xs: 12 }}>
                <Divider component="div" role="presentation">
                    <Typography sx={{ textTransform: "uppercase" }}>{translate("or")}</Typography>
                </Divider>
            </Grid>
            <Grid size={{ xs: 12 }}>
                <Button
                    id="passkey-sign-in-button"
                    variant="contained"
                    color="primary"
                    fullWidth
                    onClick={handleSignIn}
                    startIcon={<PasskeyIcon />}
                    disabled={props.disabled}
                    data-1p-ignore
                    endIcon={loading ? <CircularProgress size={20} /> : null}
                >
                    {translate("Sign in with a passkey")}
                </Button>
            </Grid>
        </Fragment>
    );
};

export default PasskeyForm;
