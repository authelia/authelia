import React, { MutableRefObject, useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Alert, AlertTitle, Button, FormControl } from "@mui/material";
import Grid from "@mui/material/Grid2";
import TextField from "@mui/material/TextField";
import { BroadcastChannel } from "broadcast-channel";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext.ts";
import { useRedirector } from "@hooks/Redirector.ts";
import { useWorkflow } from "@hooks/Workflow.ts";
import LoginLayout from "@layouts/LoginLayout.tsx";
import { UserInfo } from "@models/UserInfo.ts";
import { IsCapsLockModified } from "@services/CapsLock.ts";
import { postFirstFactorReauthenticate } from "@services/FirstFactor.ts";
import { AutheliaState } from "@services/State.ts";

export interface Props {
    userInfo: UserInfo;
    state: AutheliaState;
}

const OpenIDConnectConsentLoginFormView: React.FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation();

    const [password, setPassword] = useState("");
    const [error, setError] = useState(false);
    const [hasCapsLock, setHasCapsLock] = useState(false);
    const [isCapsLockPartial, setIsCapsLockPartial] = useState(false);
    const [disabled, setDisabled] = useState(false);
    const [workflow, workflowID] = useWorkflow();

    const redirector = useRedirector();
    const { createErrorNotification } = useNotifications();
    const loginChannel = useMemo(() => new BroadcastChannel<boolean>("login"), []);

    const passwordRef = useRef() as MutableRefObject<HTMLInputElement>;

    useEffect(() => {
        const timeout = setTimeout(() => passwordRef.current.focus(), 10);
        return () => clearTimeout(timeout);
    }, [passwordRef]);

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

        setDisabled(true);

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
            setDisabled(false);
            passwordRef.current.focus();
        }
    }, [createErrorNotification, loginChannel, password, redirector, translate, workflow, workflowID]);

    const handlePasswordKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                event.preventDefault();

                if (!password.length) {
                    passwordRef.current.focus();
                } else {
                    handleConfirm().catch(console.error);
                }
            }
        },
        [handleConfirm, password.length],
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
                            disabled={disabled}
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
                            fullWidth
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
