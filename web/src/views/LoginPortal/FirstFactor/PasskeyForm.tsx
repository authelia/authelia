import { Fragment, useCallback, useState } from "react";

import { Button, CircularProgress, Divider, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import axios from "axios";
import { useTranslation } from "react-i18next";

import PasskeyIcon from "@components/PasskeyIcon";
import { RedirectionURL, RequestMethod } from "@constants/SearchParams";
import { useAbortSignal } from "@hooks/Abort";
import { useFlow } from "@hooks/Flow";
import { useQueryParam } from "@hooks/QueryParam";
import { AssertionResult, AssertionResultFailureString } from "@models/WebAuthn";
import { getWebAuthnPasskeyOptions, getWebAuthnResult, postWebAuthnPasskeyResponse } from "@services/WebAuthn";

export interface Props {
    disabled: boolean;
    rememberMe: boolean;

    onAuthenticationStart: () => void;
    onAuthenticationStop: () => void;
    onAuthenticationError: (_err: Error) => void;
    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
}

export default function PasskeyForm(props: Props) {
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const requestMethod = useQueryParam(RequestMethod);
    const { flow, id: flowID, subflow } = useFlow();
    const getSignal = useAbortSignal();

    const [loading, setLoading] = useState(false);

    const handleSignIn = useCallback(async () => {
        if (loading) return;

        const startUI = () => {
            props.onAuthenticationStart();
            setLoading(true);
        };

        const stopUI = () => {
            props.onAuthenticationStop();
            setLoading(false);
        };

        const fail = (message: string) => {
            stopUI();
            props.onAuthenticationError(new Error(translate(message)));
        };

        startUI();

        const signal = getSignal();

        try {
            const optionsStatus = await getWebAuthnPasskeyOptions(signal);

            if (signal.aborted) return;

            if (optionsStatus.status !== 200 || optionsStatus.options == null) {
                fail("Failed to initiate security key sign in process");
                return;
            }

            const result = await getWebAuthnResult(optionsStatus.options);

            if (signal.aborted) return;

            if (result.result !== AssertionResult.Success) {
                fail(AssertionResultFailureString(result.result));
                return;
            }

            if (result.response == null) {
                fail("The browser did not respond with the expected attestation data");
                return;
            }

            const response = await postWebAuthnPasskeyResponse(
                result.response,
                props.rememberMe,
                redirectionURL,
                requestMethod,
                flowID,
                flow,
                subflow,
                signal,
            );

            stopUI();

            if (response.data.status === "OK" && response.status === 200) {
                props.onAuthenticationSuccess(response.data.data ? response.data.data.redirect : undefined);
                return;
            }

            props.onAuthenticationError(new Error(translate("The server rejected the security key")));
        } catch (err) {
            if (axios.isCancel(err)) return;
            console.error(err);
            fail("Failed to initiate security key sign in process");
        }
    }, [loading, props, getSignal, translate, redirectionURL, requestMethod, flowID, flow, subflow]);

    return (
        <Fragment>
            <Grid size={{ xs: 12 }}>
                <Divider component="div">
                    <Typography sx={{ textTransform: "uppercase" }}>{translate("or")}</Typography>
                </Divider>
            </Grid>
            <Grid size={{ xs: 12 }}>
                <Button
                    id="passkey-sign-in-button"
                    variant="contained"
                    color="primary"
                    fullWidth
                    onClick={() => handleSignIn()}
                    startIcon={<PasskeyIcon />}
                    disabled={props.disabled}
                    endIcon={loading ? <CircularProgress size={20} /> : null}
                >
                    {translate("Sign in with a passkey")}
                </Button>
            </Grid>
        </Fragment>
    );
}
