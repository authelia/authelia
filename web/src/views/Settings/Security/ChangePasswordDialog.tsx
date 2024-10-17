import React, { MutableRefObject, useCallback, useEffect, useRef, useState } from "react";

import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    Grid2,
    TextField,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import PasswordMeter from "@components/PasswordMeter";
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
    const { t: translate } = useTranslation("settings");

    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const [formDisabled, setFormDisabled] = useState(true);
    const [oldPassword, setOldPassword] = useState("");
    const [oldPasswordError, setOldPasswordError] = useState(false);
    const [newPassword, setNewPassword] = useState("");
    const [newPasswordError, setNewPasswordError] = useState(false);
    const [repeatNewPassword, setRepeatNewPassword] = useState("");
    const [repeatNewPasswordError, setRepeatNewPasswordError] = useState(false);
    const [isCapsLockOnOldPW, setIsCapsLockOnOldPW] = useState(false);
    const [isCapsLockOnNewPW, setIsCapsLockOnNewPW] = useState(false);
    const [isCapsLockOnRepeatNewPW, setIsCapsLockOnRepeatNewPW] = useState(false);

    const oldPasswordRef = useRef() as MutableRefObject<HTMLInputElement>;
    const newPasswordRef = useRef() as MutableRefObject<HTMLInputElement>;
    const repeatNewPasswordRef = useRef() as MutableRefObject<HTMLInputElement>;

    const [pPolicy, setPPolicy] = useState<PasswordPolicyConfiguration>({
        max_length: 0,
        min_length: 8,
        min_score: 0,
        require_lowercase: false,
        require_number: false,
        require_special: false,
        require_uppercase: false,
        mode: PasswordPolicyMode.Disabled,
    });

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
        setFormDisabled(false);
    }, []);

    const handleClose = useCallback(() => {
        (async () => {
            props.setClosed();
            resetStates();
        })();
    }, [props, resetStates]);

    const asyncProcess = useCallback(async () => {
        try {
            setFormDisabled(true);
            const policy = await getPasswordPolicyConfiguration();
            setPPolicy(policy);
            setFormDisabled(false);
        } catch (err) {
            console.error(err);
            createErrorNotification(
                translate("There was an issue completing the process the verification token might have expired"),
            );
            setFormDisabled(true);
        }
    }, [createErrorNotification, translate]);

    useEffect(() => {
        asyncProcess();
    }, [asyncProcess]);

    const handlePasswordChange = useCallback(async () => {
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
            return;
        }
        if (newPassword !== repeatNewPassword) {
            setNewPasswordError(true);
            setRepeatNewPasswordError(true);
            createErrorNotification(translate("Passwords do not match"));
            return;
        }

        try {
            await postPasswordChange(props.username, oldPassword, newPassword);
            createSuccessNotification(translate("Password changed successfully"));
            handleClose();
        } catch (err) {
            setOldPasswordError(false);
            setNewPasswordError(false);
            setRepeatNewPasswordError(false);
            if ((err as Error).message.includes("0000052D.")) {
                setNewPasswordError(true);
                setRepeatNewPasswordError(true);
                createErrorNotification(
                    translate("Your supplied password does not meet the password policy requirements"),
                );
            } else if ((err as Error).message.includes("policy")) {
                setNewPasswordError(true);
                setRepeatNewPasswordError(true);
                createErrorNotification(
                    translate("Your supplied password does not meet the password policy requirements"),
                );
            } else if ((err as Error).message.includes("Incorrect")) {
                setOldPasswordError(true);
                createErrorNotification(translate("Incorrect password"));
            } else if ((err as Error).message.includes("reuse")) {
                createErrorNotification(translate("You cannot reuse your old password"));
            } else {
                createErrorNotification(translate("There was an issue changing the password"));
            }
        }
    }, [
        createErrorNotification,
        createSuccessNotification,
        handleClose,
        newPassword,
        oldPassword,
        repeatNewPassword,
        props.username,
        translate,
    ]);

    const handleOldPWKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                if (!oldPassword.length) {
                    setOldPasswordError(true);
                } else if (oldPassword.length && newPassword.length && repeatNewPassword.length) {
                    handlePasswordChange().catch(console.error);
                } else {
                    setOldPasswordError(false);
                    newPasswordRef.current.focus();
                }
            }
        },
        [handlePasswordChange, oldPassword.length, newPassword.length, repeatNewPassword.length],
    );

    const handleNewPWKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                if (!newPassword.length) {
                    setNewPasswordError(true);
                } else if (oldPassword.length && newPassword.length && repeatNewPassword.length) {
                    handlePasswordChange().catch(console.error);
                } else {
                    setNewPasswordError(false);
                    repeatNewPasswordRef.current.focus();
                }
            }
        },
        [handlePasswordChange, oldPassword.length, newPassword.length, repeatNewPassword.length],
    );

    const handleRepeatNewPWKeyDown = useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                if (!repeatNewPassword.length) {
                    setNewPasswordError(true);
                } else if (oldPassword.length && newPassword.length && repeatNewPassword.length) {
                    handlePasswordChange().catch(console.error);
                } else {
                    setNewPasswordError(false);
                    repeatNewPasswordRef.current.focus();
                }
            }
        },
        [handlePasswordChange, oldPassword.length, newPassword.length, repeatNewPassword.length],
    );

    const checkCapsLockOldPW = useCallback((event: React.KeyboardEvent<HTMLDivElement>) => {
        if (event.getModifierState("CapsLock")) {
            setIsCapsLockOnOldPW(true);
        } else {
            setIsCapsLockOnOldPW(false);
        }
    }, []);

    const checkCapsLockNewPW = useCallback((event: React.KeyboardEvent<HTMLDivElement>) => {
        if (event.getModifierState("CapsLock")) {
            setIsCapsLockOnNewPW(true);
        } else {
            setIsCapsLockOnNewPW(false);
        }
    }, []);

    const checkCapsLockRepeatNewPW = useCallback((event: React.KeyboardEvent<HTMLDivElement>) => {
        if (event.getModifierState("CapsLock")) {
            setIsCapsLockOnRepeatNewPW(true);
        } else {
            setIsCapsLockOnRepeatNewPW(false);
        }
    }, []);
    const disabled = props.disabled || false;

    return (
        <Dialog open={props.open} maxWidth="xs">
            <DialogTitle>{translate("Change {{item}}", { item: translate("Password") })}</DialogTitle>
            <DialogContent>
                <FormControl id={"change-password-form"} disabled={formDisabled}>
                    <Grid2 container spacing={1} alignItems={"center"} justifyContent={"center"} textAlign={"center"}>
                        <Grid2 size={{ xs: 12 }} sx={{ pt: 3 }}>
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
                                onChange={(v) => setOldPassword(v.target.value)}
                                onFocus={() => setOldPasswordError(false)}
                                type="password"
                                autoCapitalize="none"
                                autoComplete="none"
                                onKeyDown={handleOldPWKeyDown}
                                onKeyUp={checkCapsLockOldPW}
                                helperText={isCapsLockOnOldPW ? translate("Caps Lock is on") : " "}
                                color={isCapsLockOnOldPW ? "error" : "primary"}
                                onBlur={() => setIsCapsLockOnOldPW(false)}
                            />
                        </Grid2>
                        <Grid2 size={{ xs: 12 }} sx={{ mt: 3 }}>
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
                                onChange={(v) => setNewPassword(v.target.value)}
                                onFocus={() => setNewPasswordError(false)}
                                type="password"
                                autoCapitalize="none"
                                autoComplete="none"
                                onKeyDown={handleNewPWKeyDown}
                                onKeyUp={checkCapsLockNewPW}
                                helperText={isCapsLockOnNewPW ? translate("Caps Lock is on") : " "}
                                color={isCapsLockOnNewPW ? "error" : "primary"}
                                onBlur={() => setIsCapsLockOnNewPW(false)}
                            />
                            {pPolicy.mode === PasswordPolicyMode.Disabled ? null : (
                                <PasswordMeter value={newPassword} policy={pPolicy} />
                            )}
                        </Grid2>
                        <Grid2 size={{ xs: 12 }}>
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
                                onChange={(v) => setRepeatNewPassword(v.target.value)}
                                onFocus={() => setRepeatNewPasswordError(false)}
                                type="password"
                                autoCapitalize="none"
                                autoComplete="none"
                                onKeyDown={handleRepeatNewPWKeyDown}
                                onKeyUp={checkCapsLockRepeatNewPW}
                                helperText={isCapsLockOnRepeatNewPW ? translate("Caps Lock is ON") : " "}
                                color={isCapsLockOnRepeatNewPW ? "error" : "primary"}
                                onBlur={() => setIsCapsLockOnRepeatNewPW(false)}
                            />
                        </Grid2>
                    </Grid2>
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
                    disabled={!(oldPassword.length && newPassword.length && repeatNewPassword.length)}
                >
                    {translate("Submit")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default ChangePasswordDialog;
