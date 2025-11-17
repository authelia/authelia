import React from "react";

import { Divider, Paper, Stack, Typography, useTheme } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";

import HomeButton from "@components/HomeButton";
import {
    Decision,
    ErrorDebug,
    ErrorDescription,
    ErrorHint,
    Error as ErrorParam,
    ErrorStatusCode,
    ErrorURI,
} from "@constants/SearchParams";
import LoginLayout from "@layouts/LoginLayout";
import { UserInfo } from "@models/UserInfo";
import { AutheliaState } from "@services/State";

export interface Props {
    userInfo?: UserInfo;
    state: AutheliaState;
}

const CompletionView: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent"]);
    const theme = useTheme();

    const [query] = useSearchParams();

    const decision = query.get(Decision);
    const error = query.get(ErrorParam);
    const error_description = query.get(ErrorDescription);
    const error_hint = query.get(ErrorHint);
    const error_debug = query.get(ErrorDebug);
    const error_status_code = query.get(ErrorStatusCode);
    const error_uri = query.get(ErrorURI);

    const title = error
        ? "An error occurred processing the request"
        : decision && decision === "accepted"
          ? "Consent has been accepted and processed"
          : "Consent has been rejected and processed";

    return (
        <LoginLayout id={"openid-completion-stage"} title={translate(title)} maxWidth={"sm"}>
            <Stack justifyContent={"center"} alignItems={"center"} spacing={theme.spacing(2)}>
                <HomeButton />
                {error ? (
                    <CompletionErrorView
                        error={error}
                        error_description={error_description}
                        error_hint={error_hint}
                        error_debug={error_debug}
                        error_status_code={error_status_code}
                        error_uri={error_uri}
                    />
                ) : null}

                <Typography variant={"subtitle2"} margin={theme.spacing(2)}>
                    {translate("You may close this tab or return home by clicking the home button")}.
                </Typography>
            </Stack>
        </LoginLayout>
    );
};

export default CompletionView;

interface ErrorProps {
    error: string;
    error_description: null | string;
    error_hint: null | string;
    error_debug: null | string;
    error_status_code: null | string;
    error_uri: null | string;
}
const CompletionErrorView: React.FC<ErrorProps> = (props: ErrorProps) => {
    const { t: translate } = useTranslation(["consent"]);
    const theme = useTheme();

    return (
        <Paper sx={{ padding: theme.spacing(2) }} elevation={24}>
            <Stack spacing={theme.spacing(2)}>
                <Typography variant={"h6"}>
                    <strong>{translate("Error")}:</strong> {translate(props.error)}
                </Typography>
                {props.error_description || props.error_hint || props.error_debug || props.error_uri ? (
                    <Divider />
                ) : null}
                {props.error_description ? (
                    <Typography>
                        <strong>{translate("Description")}:</strong> {translate(props.error_description)}
                    </Typography>
                ) : null}
                {props.error_hint ? (
                    <Typography>
                        <strong>{translate("Hint")}:</strong> {translate(props.error_hint)}
                    </Typography>
                ) : null}
                {props.error_debug ? (
                    <Typography>
                        <strong>{translate("Debug Information")}:</strong> {translate(props.error_debug)}
                    </Typography>
                ) : null}
                {props.error_uri ? (
                    <Typography>
                        <strong>{translate("Documentation")}:</strong> {translate(props.error_uri)}
                    </Typography>
                ) : null}
            </Stack>
        </Paper>
    );
};
