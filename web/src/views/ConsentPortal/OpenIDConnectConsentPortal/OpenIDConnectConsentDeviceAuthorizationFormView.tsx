import React, { useEffect, useState } from "react";

import { Button, FormControl, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
import { useUserCode } from "@hooks/OpenIDConnect";
import LoginLayout from "@layouts/LoginLayout";
import { UserInfo } from "@models/UserInfo";
import { AutheliaState } from "@services/State";

export interface Props {
    userInfo: UserInfo;
    state: AutheliaState;
}

const OpenIDConnectConsentDeviceAuthorizationFormView: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation();
    const theme = useTheme();

    const [code, setCode] = useState("");

    const userCode = useUserCode();

    useEffect(() => {
        if (userCode === null || userCode === "") {
            return;
        }

        setCode(userCode);
    }, [userCode]);

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
                                <Button id="confirm-button" variant="contained" color="primary" fullWidth>
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
