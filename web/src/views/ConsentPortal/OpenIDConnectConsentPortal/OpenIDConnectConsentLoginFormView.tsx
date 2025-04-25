import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Alert, AlertTitle, Button, CircularProgress, FormControl, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { BroadcastChannel } from "broadcast-channel";
import { useTranslation } from "react-i18next";

import LogoutButton from "@components/LogoutButton";
import { useFlow } from "@hooks/Flow";
import { useNotifications } from "@hooks/NotificationsContext";
import { useRedirector } from "@hooks/Redirector";
import LoginLayout from "@layouts/LoginLayout";
import { UserInfo } from "@models/UserInfo";
import { IsCapsLockModified } from "@services/CapsLock";
import { postFirstFactorReauthenticate } from "@services/Password";
import { AutheliaState } from "@services/State";

export interface Props {
    userInfo: UserInfo;
    state: AutheliaState;
}

const OpenIDConnectConsentLoginFormView: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation();
    const theme = useTheme();

    const [password, setPassword] = useState("");
    const [error, setError] = useState(false);
    const [hasCapsLock, setHasCapsLock] = useState(false);
    const [isCapsLockPartial, setIsCapsLockPartial] = useState(false);
    const [loading, setLoading] = useState(false);
    const { id: flowID, flow, subflow } = useFlow();

    const redirector = useRedirector();
    const { createErrorNotification } = useNotifications();
    const loginChannel = useMemo(() => new BroadcastChannel<boolean>("login"), []);

    const passwordRef = useRef<HTMLInputElement | null>(null);

    const focusPassword = useCallback(() => {
        if (passwordRef.current === null) return;

        passwordRef.current.focus();
    }, [passwordRef]);

    useEffect(() => {
        const timeout = setTimeout(() => focusPassword(), 10);
        return () => clearTimeout(timeout);
    }, [focusPassword]);

    const handleConfirm = useCallback(async () => {
        if (password === "") {
            if (password === "") {
                setError(true);
            }

            return;
        }

        if (flow === undefined || flowID === undefined) {
            return;
        }

        setLoading(true);

        try {
            const res = await postFirstFactorReauthenticate(password, undefined, undefined, flowID, flow, subflow);
            await loginChannel.postMessage(true);

            if (res) {
                redirector(res.redirect);
            }
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Failed to confirm your identity"));
            setPassword("");
            setLoading(false);
            focusPassword();
        }
    }, [password, flow, flowID, subflow, loginChannel, redirector, createErrorNotification, translate, focusPassword]);

    const handlePasswordKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                event.preventDefault();

                if (!password.length) {
                    focusPassword();
                } else {
                    handleConfirm().catch(console.error);
                }
            }
        },
        [focusPassword, handleConfirm, password.length],
    );

    const handlePasswordKeyUp = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (password.length <= 1) {
                setHasCapsLock(false);
                setIsCapsLockPartial(false);

                if (password.length === 0) {
                    return;
                }
            }

            const modified = IsCapsLockModified(event);

            if (modified === null) return;

            if (modified) {
                setHasCapsLock(true);
            } else {
                setIsCapsLockPartial(true);
            }
        },
        [password.length],
    );

    return (
        <LoginLayout id="consent-stage" title={translate("Confirm Access")}>
            <Grid container direction={"column"} justifyContent={"center"} alignItems={"center"}>
                <Grid size={{ xs: 12 }} sx={{ paddingBottom: theme.spacing(2) }}>
                    <LogoutButton />
                </Grid>
                <Grid size={{ xs: 12 }}>
                    <FormControl id={"form-consent-openid-device-code-authorization"}>
                        <Grid container spacing={2}>
                            <Grid size={{ xs: 12 }}>
                                <TextField
                                    id="password-textfield"
                                    label={translate("Password")}
                                    variant="outlined"
                                    inputRef={passwordRef}
                                    onKeyDown={handlePasswordKeyDown}
                                    onKeyUp={handlePasswordKeyUp}
                                    error={error}
                                    disabled={loading}
                                    value={password}
                                    onChange={(v) => setPassword(v.target.value)}
                                    onFocus={() => setError(false)}
                                    type="password"
                                    autoComplete="current-password"
                                    required
                                    fullWidth
                                />
                            </Grid>
                            {hasCapsLock ? (
                                <Grid size={{ xs: 12 }} marginX={2}>
                                    <Alert severity={"warning"}>
                                        <AlertTitle>{translate("Warning")}</AlertTitle>
                                        {isCapsLockPartial
                                            ? translate("The password was partially entered with Caps Lock")
                                            : translate("The password was entered with Caps Lock")}
                                    </Alert>
                                </Grid>
                            ) : null}
                            <Grid size={{ xs: 12 }}>
                                <Button
                                    id="confirm-button"
                                    variant="contained"
                                    color="primary"
                                    fullWidth={true}
                                    endIcon={loading ? <CircularProgress size={20} /> : null}
                                    disabled={loading}
                                    onClick={handleConfirm}
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

export default OpenIDConnectConsentLoginFormView;
