import React from "react";

import { Grid, Typography, useTheme } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";

import HomeButton from "@components/HomeButton";
import { Decision, ErrorDescription, ErrorHint, Error as ErrorParam, ErrorURI } from "@constants/SearchParams";
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
    const error_uri = query.get(ErrorURI);

    const title = error
        ? "An error occurred processing the request"
        : decision && decision === "accepted"
          ? "Consent has been accepted and processed"
          : "Consent has been rejected and processed";

    return (
        <LoginLayout id={"openid-completion-stage"} title={translate(title)}>
            <Grid container direction={"column"} justifyContent={"center"} alignItems={"center"}>
                <Grid size={{ xs: 12 }} sx={{ paddingBottom: theme.spacing(2) }}>
                    <HomeButton />
                </Grid>
                {error ? (
                    <Grid size={{ xs: 12 }}>
                        <Typography variant={"h4"}>{`${translate("Error Code")}: ${translate(error)}`}</Typography>
                    </Grid>
                ) : null}
                {error && error_description ? (
                    <Grid size={{ xs: 12 }}>
                        <Typography
                            variant={"body1"}
                        >{`${translate("Description")}: ${translate(error_description)}`}</Typography>
                    </Grid>
                ) : null}
                {error && error_hint ? (
                    <Grid size={{ xs: 12 }}>
                        <Typography variant={"body1"}>{`${translate("Hint")}: ${translate(error_hint)}`}</Typography>
                    </Grid>
                ) : null}
                {error && error_uri ? (
                    <Grid size={{ xs: 12 }}>
                        <Typography
                            variant={"body1"}
                        >{`${translate("Documentation")}: ${translate(error_uri)}`}</Typography>
                    </Grid>
                ) : null}
                <Grid size={{ xs: 12 }}>
                    <Typography variant={"body1"}>
                        {translate("You may close this tab or return home by clicking the home button")}
                    </Typography>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default CompletionView;
