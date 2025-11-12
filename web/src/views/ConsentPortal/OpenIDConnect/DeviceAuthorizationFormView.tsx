import React, { useCallback, useEffect, useRef, useState } from "react";

import { Box, Button, FormControl, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
import SwitchUserButton from "@components/SwitchUserButton";
import { ConsentDecisionSubRoute, ConsentOpenIDSubRoute, ConsentRoute, IndexRoute } from "@constants/Routes";
import {
    Flow,
    FlowNameOpenIDConnect,
    SubFlow,
    SubFlowNameDeviceAuthorization,
    UserCode,
} from "@constants/SearchParams";
import { useUserCode } from "@hooks/OpenIDConnect";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import LoginLayout from "@layouts/LoginLayout";
import { AutheliaState, AuthenticationLevel } from "@services/State";
import LoadingPage from "@views/LoadingPage/LoadingPage";

export interface Props {
    state: AutheliaState;
}

const DeviceAuthorizationFormView: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent", "settings"]);
    const theme = useTheme();

    const userCode = useUserCode();

    const [code, setCode] = useState(userCode || "");

    const navigate = useRouterNavigate();

    const autoSubmittedRef = useRef(false);

    const handleCode = useCallback(
        (code: string) => {
            if (code === "") {
                return;
            }

            const params = new URLSearchParams();

            params.set(UserCode, code);
            params.set(Flow, FlowNameOpenIDConnect);
            params.set(SubFlow, SubFlowNameDeviceAuthorization);

            navigate(`${ConsentRoute}${ConsentOpenIDSubRoute}${ConsentDecisionSubRoute}`, true, true, true, params);
        },
        [navigate],
    );

    useEffect(() => {
        if (props.state.authentication_level === AuthenticationLevel.Unauthenticated) {
            const params = new URLSearchParams();

            if (userCode) {
                params.set(UserCode, userCode);
            }

            params.set(Flow, FlowNameOpenIDConnect);
            params.set(SubFlow, SubFlowNameDeviceAuthorization);

            navigate(IndexRoute, true, true, true, params);
        }
    }, [userCode, navigate, props.state.authentication_level]);

    useEffect(() => {
        if (
            !userCode ||
            props.state.authentication_level === AuthenticationLevel.Unauthenticated ||
            autoSubmittedRef.current
        ) {
            return;
        }

        autoSubmittedRef.current = true;
        handleCode(userCode);
    }, [handleCode, props.state.authentication_level, userCode]);

    return props.state.authentication_level === AuthenticationLevel.Unauthenticated ? (
        <Box>
            <LoadingPage />
        </Box>
    ) : (
        <LoginLayout id={"openid-consent-device-auth-stage"} title={translate("Confirm the Code")}>
            <Grid container direction={"column"} justifyContent={"center"} alignItems={"center"}>
                <Grid size={{ xs: 12 }} sx={{ paddingBottom: theme.spacing(2) }}>
                    <LogoutButton /> {" | "} <SwitchUserButton />
                </Grid>
                <Grid size={{ xs: 12 }}>
                    <FormControl id={"form-consent-openid-device-code-authorization"}>
                        <Grid container spacing={2}>
                            <Grid size={{ xs: 12 }}>
                                <TextField
                                    id={"user-code"}
                                    label={translate("Code")}
                                    variant={"outlined"}
                                    required
                                    value={code}
                                    fullWidth
                                    onChange={(v) => setCode(v.target.value)}
                                    autoCapitalize={"none"}
                                />
                            </Grid>
                            <Grid size={{ xs: 12 }}>
                                <Button
                                    id={"confirm-button"}
                                    variant={"contained"}
                                    color={"primary"}
                                    fullWidth
                                    onClick={() => handleCode(code)}
                                    disabled={code === ""}
                                >
                                    {translate("Confirm", { ns: "settings" })}
                                </Button>
                            </Grid>
                        </Grid>
                    </FormControl>
                </Grid>
            </Grid>
        </LoginLayout>
    );
};

export default DeviceAuthorizationFormView;
