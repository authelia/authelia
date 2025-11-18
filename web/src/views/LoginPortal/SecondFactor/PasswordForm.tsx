import { KeyboardEvent, useCallback, useEffect, useRef, useState } from "react";

import Visibility from "@mui/icons-material/Visibility";
import VisibilityOff from "@mui/icons-material/VisibilityOff";
import { Alert, AlertTitle, Button, CircularProgress, FormControl, IconButton, InputAdornment } from "@mui/material";
import Grid from "@mui/material/Grid";
import TextField from "@mui/material/TextField";
import { useTranslation } from "react-i18next";

import { RedirectionURL } from "@constants/SearchParams";
import { useFlow } from "@hooks/Flow";
import { useNotifications } from "@hooks/NotificationsContext";
import { useQueryParam } from "@hooks/QueryParam";
import { IsCapsLockModified } from "@services/CapsLock";
import { postSecondFactor } from "@services/Password";

export interface Props {
    onAuthenticationSuccess: (_redirectURL: string | undefined) => void;
}

const PasswordForm = function (props: Props) {
    const { createErrorNotification } = useNotifications();
    const { t: translate } = useTranslation(["portal", "settings"]);

    const redirectionURL = useQueryParam(RedirectionURL);
    const { flow, id: flowID, subflow } = useFlow();

    const [loading, setLoading] = useState(false);
    const [password, setPassword] = useState("");
    const [passwordCapsLock, setPasswordCapsLock] = useState(false);
    const [passwordCapsLockPartial, setPasswordCapsLockPartial] = useState(false);
    const [passwordError, setPasswordError] = useState(false);
    const [showPassword, setShowPassword] = useState(false);

    const passwordRef = useRef<HTMLInputElement | null>(null);

    const focusPassword = useCallback(() => {
        if (passwordRef.current === null) return;

        passwordRef.current.focus();
    }, [passwordRef]);

    useEffect(() => {
        const timeout = setTimeout(() => focusPassword(), 10);
        return () => clearTimeout(timeout);
    }, [focusPassword]);

    const handleSignIn = useCallback(async () => {
        if (password === "") {
            setPasswordError(true);

            return;
        }

        setLoading(true);

        try {
            const res = await postSecondFactor(password, redirectionURL, flowID, flow, subflow);
            props.onAuthenticationSuccess(res ? res.redirect : undefined);
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Incorrect password"));
            setPassword("");
            setLoading(false);
            focusPassword();
        }
    }, [createErrorNotification, focusPassword, password, props, redirectionURL, translate, flowID, flow, subflow]);

    const handlePasswordKeyDown = useCallback(
        (event: KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                if (!password.length) {
                    focusPassword();
                }
                handleSignIn().catch(console.error);
                event.preventDefault();
            }
        },
        [focusPassword, handleSignIn, password.length],
    );

    const handlePasswordKeyUp = useCallback(
        (event: KeyboardEvent<HTMLDivElement>) => {
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
                        type={showPassword ? "text" : "password"}
                        autoComplete="current-password"
                        onKeyDown={handlePasswordKeyDown}
                        onKeyUp={handlePasswordKeyUp}
                        slotProps={{
                            input: {
                                endAdornment: (
                                    <InputAdornment position="end">
                                        <IconButton
                                            aria-label={translate("Toggle password visibility")}
                                            edge="end"
                                            size="large"
                                            onMouseDown={() => setShowPassword(true)}
                                            onMouseUp={() => setShowPassword(false)}
                                            onMouseLeave={() => setShowPassword(false)}
                                            onTouchStart={() => setShowPassword(true)}
                                            onTouchEnd={() => setShowPassword(false)}
                                            onTouchCancel={() => setShowPassword(false)}
                                            onKeyDown={(e) => {
                                                if (e.key === " ") {
                                                    setShowPassword(true);
                                                    e.preventDefault();
                                                }
                                            }}
                                            onKeyUp={(e) => {
                                                if (e.key === " ") {
                                                    setShowPassword(false);
                                                    e.preventDefault();
                                                }
                                            }}
                                        >
                                            {showPassword ? <Visibility /> : <VisibilityOff />}
                                        </IconButton>
                                    </InputAdornment>
                                ),
                            },
                        }}
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
                        fullWidth={true}
                        endIcon={loading ? <CircularProgress size={20} /> : null}
                        disabled={loading}
                        onClick={handleSignIn}
                    >
                        {translate("Authenticate", { ns: "settings" })}
                    </Button>
                </Grid>
            </Grid>
        </FormControl>
    );
};

export default PasswordForm;
