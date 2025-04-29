import React, { useCallback, useEffect, useState } from "react";

import { Button, FormControl, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
import {
    FlowNameOpenIDConnect,
    QueryParamFlow,
    QueryParamSubFlow,
    SubFlowNameDeviceAuthorization,
} from "@constants/constants";
import { ConsentDecisionSubRoute, ConsentOpenIDSubRoute, ConsentRoute, IndexRoute } from "@constants/Routes";
import { QueryParamUserCode, useUserCode } from "@hooks/OpenIDConnect";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import LoginLayout from "@layouts/LoginLayout";
import { AutheliaState, AuthenticationLevel } from "@services/State";

export interface Props {
    state: AutheliaState;
}

const OpenIDConnectConsentDeviceAuthorizationFormView: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation();
    const theme = useTheme();

    const [code, setCode] = useState("");

    const userCode = useUserCode();

    const navigate = useRouterNavigate();

    const handleCode = useCallback(
        (code: string) => {
            if (code === "") {
                return;
            }

            const params = new URLSearchParams();

            params.set(QueryParamUserCode, code);
            params.set(QueryParamFlow, FlowNameOpenIDConnect);
            params.set(QueryParamSubFlow, SubFlowNameDeviceAuthorization);

            navigate(`${ConsentRoute}${ConsentOpenIDSubRoute}${ConsentDecisionSubRoute}`, true, true, true, params);
        },
        [navigate],
    );

    useEffect(() => {
        if (props.state.authentication_level === AuthenticationLevel.Unauthenticated) {
            const params = new URLSearchParams();

            if (code) {
                params.set(QueryParamUserCode, code);
            }

            params.set(QueryParamFlow, FlowNameOpenIDConnect);
            params.set(QueryParamSubFlow, SubFlowNameDeviceAuthorization);

            navigate(IndexRoute, true, true, true, params);
        }
    }, [code, navigate, props.state.authentication_level]);

    useEffect(() => {
        if (!userCode || props.state.authentication_level === AuthenticationLevel.Unauthenticated) {
            return;
        }

        setCode(userCode);
        handleCode(userCode);
    }, [handleCode, props.state.authentication_level, userCode]);

    return (
        <LoginLayout id="consent-stage" title={translate("Confirm the Code")}>
            <Grid container direction={"column"} justifyContent={"center"} alignItems={"center"}>
                <Grid size={{ xs: 12 }} sx={{ paddingBottom: theme.spacing(2) }}>
                    <LogoutButton />
                </Grid>
                <Grid size={{ xs: 12 }}>
                    <FormControl id={"form-consent-openid-device-code-authorization"}>
                        <Grid container spacing={2}>
                            <Grid size={{ xs: 12 }}>
                                <TextField
                                    id="user-code"
                                    label={translate("Code")}
                                    variant="outlined"
                                    required
                                    value={code}
                                    fullWidth
                                    onChange={(v) => setCode(v.target.value)}
                                    autoCapitalize="none"
                                />
                            </Grid>
                            <Grid size={{ xs: 12 }}>
                                <Button
                                    id="confirm-button"
                                    variant="contained"
                                    color="primary"
                                    fullWidth
                                    onClick={() => handleCode(code)}
                                    disabled={code === ""}
                                >
                                    {translate("Confirm")}
                                </Button>
                            </Grid>
                        </Grid>
                    </FormControl>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default OpenIDConnectConsentDeviceAuthorizationFormView;
