import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Button, CircularProgress, FormControl } from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { BroadcastChannel } from "broadcast-channel";
import { useTranslation } from "react-i18next";

import useCheckCapsLock from "@hooks/CapsLock";
import { useNotifications } from "@hooks/NotificationsContext";
import { useRedirector } from "@hooks/Redirector";
import { useWorkflow } from "@hooks/Workflow";
import LoginLayout from "@layouts/LoginLayout";
import { UserInfo } from "@models/UserInfo";
import { postFirstFactorReauthenticate } from "@services/Password";
import { AutheliaState } from "@services/State";

export interface Props {
    userInfo: UserInfo;
    state: AutheliaState;
}

const OpenIDConnectConsentLoginFormView: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation();

    const [password, setPassword] = useState("");
    const [error, setError] = useState(false);
    const [passwordCapsLock, setPasswordCapsLock] = useState(false);
    const [loading, setLoading] = useState(false);
    const [workflow, workflowID] = useWorkflow();

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

        if (workflow === undefined || workflowID === undefined) {
            return;
        }

        setLoading(true);

        try {
            const res = await postFirstFactorReauthenticate(password, undefined, undefined, workflow, workflowID);
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
    }, [createErrorNotification, focusPassword, loginChannel, password, redirector, translate, workflow, workflowID]);

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

    return (
        <LoginLayout id="consent-stage" title={translate("Confirm Access")}>
            <FormControl id={"form-consent-openid-device-code-authorization"}>
                <Grid container spacing={2}>
                    <Grid size={{ xs: 12 }}>
                        <TextField
                            id="password-textfield"
                            label={translate("Password")}
                            variant="outlined"
                            inputRef={passwordRef}
                            onKeyDown={handlePasswordKeyDown}
                            onKeyUp={useCheckCapsLock(setPasswordCapsLock)}
                            error={error}
                            disabled={loading}
                            value={password}
                            onChange={(v) => setPassword(v.target.value)}
                            onFocus={() => setError(false)}
                            type="password"
                            autoComplete="current-password"
                            required
                            fullWidth
                            helperText={passwordCapsLock ? translate("Caps Lock is on") : ""}
                            onBlur={() => setPasswordCapsLock(false)}
                            color={passwordCapsLock ? "warning" : "primary"}
                        />
                    </Grid>
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
        </LoginLayout>
    );
};

export default OpenIDConnectConsentLoginFormView;
