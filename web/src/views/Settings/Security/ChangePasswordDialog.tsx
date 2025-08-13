import { Dispatch, KeyboardEvent, RefObject, SetStateAction, useCallback, useEffect, useRef, useState } from "react";

import {
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    TextField,
} from "@mui/material";
import Grid from "@mui/material/Grid";
import axios from "axios";
import { useTranslation } from "react-i18next";

import PasswordMeter from "@components/PasswordMeter";
import useCheckCapsLock from "@hooks/CapsLock";
import { useNotifications } from "@hooks/NotificationsContext";
import { PasswordPolicyConfiguration, PasswordPolicyMode } from "@models/PasswordPolicy";
import { postPasswordChange } from "@services/ChangePassword";
import { getPasswordPolicyConfiguration } from "@services/PasswordPolicyConfiguration";

interface Props {
    username: string;
    disabled?: boolean;
    open: boolean;
    setClosed: () => void;
}

const ChangePasswordDialog = (props: Props) => {
    const { t: translate } = useTranslation(["settings", "portal"]);

    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [loading, setLoading] = useState(false);
    const [oldPassword, setOldPassword] = useState("");
    const [oldPasswordError, setOldPasswordError] = useState(false);
    const [newPassword, setNewPassword] = useState("");
    const [newPasswordError, setNewPasswordError] = useState(false);
    const [repeatNewPassword, setRepeatNewPassword] = useState("");
    const [repeatNewPasswordError, setRepeatNewPasswordError] = useState(false);
    const [isCapsLockOnOldPW, setIsCapsLockOnOldPW] = useState(false);
    const [isCapsLockOnNewPW, setIsCapsLockOnNewPW] = useState(false);
    const [isCapsLockOnRepeatNewPW, setIsCapsLockOnRepeatNewPW] = useState(false);

    const oldPasswordRef = useRef<HTMLInputElement | null>(null);
    const newPasswordRef = useRef<HTMLInputElement | null>(null);
    const repeatNewPasswordRef = useRef<HTMLInputElement | null>(null);

    const [pPolicy, setPPolicy] = useState<PasswordPolicyConfiguration>({
        max_length: 0,
        min_length: 8,
        min_score: 0,
        mode: PasswordPolicyMode.Disabled,
        require_lowercase: false,
        require_number: false,
        require_special: false,
        require_uppercase: false,
    });

    const resetPasswordErrors = useCallback(() => {
        setOldPasswordError(false);
        setNewPasswordError(false);
        setRepeatNewPasswordError(false);
    }, []);

    const resetStates = useCallback(() => {
        setOldPassword("");
        setNewPassword("");
        setRepeatNewPassword("");

        setOldPasswordError(false);
        setNewPasswordError(false);
        setRepeatNewPasswordError(false);

        setIsCapsLockOnOldPW(false);
        setIsCapsLockOnNewPW(false);
        setIsCapsLockOnRepeatNewPW(false);
    }, []);

    const handleClose = useCallback(() => {
        props.setClosed();
        resetStates();
    }, [props, resetStates]);

    const asyncProcess = useCallback(async () => {
        try {
            setLoading(true);
            const policy = await getPasswordPolicyConfiguration();
            setPPolicy(policy);
        } catch {
            createErrorNotification(
                translate("There was an issue completing the process the verification token might have expired", {
                    ns: "portal",
                }),
            );
            setLoading(true);
        }
    }, [createErrorNotification, translate]);

    useEffect(() => {
        asyncProcess();
    }, [asyncProcess]);

    const handlePasswordChange = useCallback(async () => {
        setLoading(true);
        if (oldPassword.trim() === "" || newPassword.trim() === "" || repeatNewPassword.trim() === "") {
            if (oldPassword.trim() === "") {
                setOldPasswordError(true);
            }
            if (newPassword.trim() === "") {
                setNewPasswordError(true);
            }
            if (repeatNewPassword.trim() === "") {
                setRepeatNewPasswordError(true);
            }
            setLoading(false);
            return;
        }
        if (newPassword !== repeatNewPassword) {
            setNewPasswordError(true);
            setRepeatNewPasswordError(true);
            createErrorNotification(translate("Passwords do not match"));
            setLoading(false);
            return;
        }

        try {
            await postPasswordChange(props.username, oldPassword, newPassword);
            createSuccessNotification(translate("Password changed successfully"));
            handleClose();
        } catch (err) {
            resetPasswordErrors();
            setLoading(false);
            if (axios.isAxiosError(err) && err.response) {
                switch (err.response.status) {
                    case 400: // Bad Request - Weak Password
                        setNewPasswordError(true);
                        setRepeatNewPasswordError(true);
                        createErrorNotification(
                            translate("Your supplied password does not meet the password policy requirements"),
                        );
                        break;

                    case 401: // Unauthorized - Incorrect Password
                        setOldPasswordError(true);
                        createErrorNotification(translate("Incorrect password"));
                        break;

                    case 500: // Internal Server Error
                    default:
                        createErrorNotification(translate("There was an issue changing the password"));
                        break;
                }
            } else {
                // Handle non-axios errors
                createErrorNotification(translate("There was an issue changing the password"));
            }
            return;
        }
    }, [
        createErrorNotification,
        createSuccessNotification,
        resetPasswordErrors,
        handleClose,
        newPassword,
        oldPassword,
        repeatNewPassword,
        props.username,
        translate,
    ]);

    const useHandleKeyDown = (
        passwordState: string,
        setError: Dispatch<SetStateAction<boolean>>,
        nextRef?: RefObject<HTMLInputElement | null>,
    ) => {
        return useCallback(
            (event: KeyboardEvent<HTMLDivElement>) => {
                if (event.key === "Enter") {
                    if (!passwordState.length) {
                        setError(true);
                    } else if (!nextRef) {
                        handlePasswordChange().catch(console.error);
                    } else if (nextRef.current) {
                        nextRef.current.focus();
                    }
                }
            },
            [nextRef, passwordState.length, setError],
        );
    };

    const handleOldPWKeyDown = useHandleKeyDown(oldPassword, setOldPasswordError, newPasswordRef);
    const handleNewPWKeyDown = useHandleKeyDown(newPassword, setNewPasswordError, repeatNewPasswordRef);
    const handleRepeatNewPWKeyDown = useHandleKeyDown(repeatNewPassword, setRepeatNewPasswordError);

    const disabled = props.disabled || false;

    return (
        <Dialog open={props.open} maxWidth="xs">
            <DialogTitle>{translate("Change Password")}</DialogTitle>
            <DialogContent>
                <FormControl id={"change-password-form"} disabled={loading}>
                    <Grid container spacing={1} alignItems={"center"} justifyContent={"center"} textAlign={"center"}>
                        <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                            <TextField
                                inputRef={oldPasswordRef}
                                id="old-password"
                                label={translate("Old Password")}
                                variant="outlined"
                                required
                                value={oldPassword}
                                error={oldPasswordError}
                                disabled={disabled}
                                fullWidth
                                onChange={(v: { target: { value: React.SetStateAction<string> } }) =>
                                    setOldPassword(v.target.value)
                                }
                                onFocus={() => setOldPasswordError(false)}
                                type="password"
                                autoCapitalize="off"
                                autoComplete="off"
                                onKeyDown={handleOldPWKeyDown}
                                onKeyUp={useCheckCapsLock(setIsCapsLockOnOldPW)}
                                helperText={isCapsLockOnOldPW ? translate("Caps Lock is on") : " "}
                                color={isCapsLockOnOldPW ? "error" : "primary"}
                                onBlur={() => setIsCapsLockOnOldPW(false)}
                            />
                        </Grid>
                        <Grid size={{ xs: 12 }} sx={{ mt: 3 }}>
                            <TextField
                                inputRef={newPasswordRef}
                                id="new-password"
                                label={translate("New Password")}
                                variant="outlined"
                                required
                                fullWidth
                                disabled={disabled}
                                value={newPassword}
                                error={newPasswordError}
                                onChange={(v: { target: { value: React.SetStateAction<string> } }) =>
                                    setNewPassword(v.target.value)
                                }
                                onFocus={() => setNewPasswordError(false)}
                                type="password"
                                autoCapitalize="off"
                                autoComplete="off"
                                onKeyDown={handleNewPWKeyDown}
                                onKeyUp={useCheckCapsLock(setIsCapsLockOnNewPW)}
                                helperText={isCapsLockOnNewPW ? translate("Caps Lock is on") : " "}
                                color={isCapsLockOnNewPW ? "error" : "primary"}
                                onBlur={() => setIsCapsLockOnNewPW(false)}
                            />
                            {pPolicy.mode === PasswordPolicyMode.Disabled ? null : (
                                <PasswordMeter value={newPassword} policy={pPolicy} />
                            )}
                        </Grid>
                        <Grid size={{ xs: 12 }}>
                            <TextField
                                inputRef={repeatNewPasswordRef}
                                id="repeat-new-password"
                                label={translate("Repeat New Password")}
                                variant="outlined"
                                required
                                fullWidth
                                disabled={disabled}
                                value={repeatNewPassword}
                                error={repeatNewPasswordError}
                                onChange={(v: { target: { value: React.SetStateAction<string> } }) =>
                                    setRepeatNewPassword(v.target.value)
                                }
                                onFocus={() => setRepeatNewPasswordError(false)}
                                type="password"
                                autoCapitalize="off"
                                autoComplete="off"
                                onKeyDown={handleRepeatNewPWKeyDown}
                                onKeyUp={useCheckCapsLock(setIsCapsLockOnRepeatNewPW)}
                                helperText={isCapsLockOnRepeatNewPW ? translate("Caps Lock is on") : " "}
                                color={isCapsLockOnRepeatNewPW ? "error" : "primary"}
                                onBlur={() => setIsCapsLockOnRepeatNewPW(false)}
                            />
                        </Grid>
                    </Grid>
                </FormControl>
            </DialogContent>
            <DialogActions>
                <Button id={"password-change-dialog-cancel"} color={"error"} onClick={handleClose}>
                    {translate("Cancel")}
                </Button>
                <Button
                    id={"password-change-dialog-submit"}
                    color={"primary"}
                    onClick={handlePasswordChange}
                    disabled={!(oldPassword.length && newPassword.length && repeatNewPassword.length) || loading}
                    startIcon={loading ? <CircularProgress color="inherit" size={20} /> : <></>}
                >
                    {translate("Submit")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default ChangePasswordDialog;
