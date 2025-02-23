import React, { MutableRefObject, useCallback, useEffect, useRef, useState } from "react";

import { Alert, AlertTitle, Button, CircularProgress, FormControl } from "@mui/material";
import Grid from "@mui/material/Grid2";
import TextField from "@mui/material/TextField";
import { useTranslation } from "react-i18next";

import { RedirectionURL } from "@constants/SearchParams";
import { useNotifications } from "@hooks/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { useWorkflow } from "@hooks/Workflow";
import { IsCapsLockModified } from "@services/CapsLock";
import { postSecondFactor } from "@services/Password";

export interface Props {
    onAuthenticationSuccess: (redirectURL: string | undefined) => void;
}

const PasswordForm = function (props: Props) {
    const { createErrorNotification } = useNotifications();
    const { t: translate } = useTranslation();

    const redirectionURL = useQueryParam(RedirectionURL);
    const [workflow, workflowID] = useWorkflow();

    const [loading, setLoading] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordCapsLock, setPasswordCapsLock] = useState(false);
    const [passwordCapsLockPartial, setPasswordCapsLockPartial] = useState(false);
    const [passwordError, setPasswordError] = useState(false);

    const passwordRef = useRef() as MutableRefObject<HTMLInputElement>;

    useEffect(() => {
        const timeout = setTimeout(() => passwordRef.current.focus(), 10);
        return () => clearTimeout(timeout);
    }, [passwordRef]);

    const handleSignIn = useCallback(async () => {
        if (password === "") {
            setPasswordError(true);

            return;
        }

        setLoading(true);

        try {
            const res = await postSecondFactor(password, redirectionURL, workflow, workflowID);
            props.onAuthenticationSuccess(res ? res.redirect : undefined);
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Incorrect password"));
            setPassword("");
            setLoading(false);
            passwordRef.current.focus();
        }
    }, [createErrorNotification, password, props, redirectionURL, translate, workflow, workflowID]);

    const handlePasswordKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                if (!password.length) {
                    passwordRef.current.focus();
                }
                handleSignIn().catch(console.error);
                event.preventDefault();
            }
        },
        [handleSignIn, password.length],
    );

    const handlePasswordKeyUp = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (password.length <= 1) {
                setPasswordCapsLock(false);
                setPasswordCapsLockPartial(false);

                if (password.length === 0) {
                    return;
                }
            }

            const modified = IsCapsLockModified(event);

            if (modified === null) return;

            if (modified) {
                setPasswordCapsLock(true);
            } else {
                setPasswordCapsLockPartial(true);
            }
        },
        [password.length],
    );

    return (
        <FormControl id={"form-password"}>
            <Grid container spacing={2}>
                <Grid size={{ xs: 12 }}>
                    <TextField
                        inputRef={passwordRef}
                        id="password-textfield"
                        label={translate("Password")}
                        variant="outlined"
                        required
                        fullWidth
                        disabled={loading}
                        value={password}
                        error={passwordError}
                        onChange={(v) => setPassword(v.target.value)}
                        onFocus={() => setPasswordError(false)}
                        type="password"
                        autoComplete="current-password"
                        onKeyDown={handlePasswordKeyDown}
                        onKeyUp={handlePasswordKeyUp}
                    />
                </Grid>
                {passwordCapsLock ? (
                    <Grid size={{ xs: 12 }} marginX={2}>
                        <Alert severity={"warning"}>
                            <AlertTitle>{translate("Warning")}</AlertTitle>
                            {passwordCapsLockPartial
                                ? translate("The password was partially entered with Caps Lock")
                                : translate("The password was entered with Caps Lock")}
                        </Alert>
                    </Grid>
                ) : null}
                <Grid size={{ xs: 12 }}>
                    <Button
                        id="sign-in-button"
                        variant="contained"
                        color="primary"
                        fullWidth
                        disabled={loading}
                        onClick={handleSignIn}
                        endIcon={loading ? <CircularProgress size={20} /> : null}
                    >
                        {translate("Authenticate")}
                    </Button>
                </Grid>
            </Grid>
        </FormControl>
    );
};

export default PasswordForm;
