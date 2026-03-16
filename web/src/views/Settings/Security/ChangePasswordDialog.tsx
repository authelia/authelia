import { Dispatch, KeyboardEvent, RefObject, SetStateAction, useCallback, useEffect, useRef, useState } from "react";

import axios from "axios";
import { useTranslation } from "react-i18next";

import PasswordMeter from "@components/PasswordMeter";
import { Button } from "@components/UI/Button";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@components/UI/Dialog";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { Spinner } from "@components/UI/Spinner";
import { useNotifications } from "@contexts/NotificationsContext";
import useCheckCapsLock from "@hooks/CapsLock";
import { PasswordPolicyConfiguration, PasswordPolicyMode } from "@models/PasswordPolicy";
import { postPasswordChange } from "@services/ChangePassword";
import { getPasswordPolicyConfiguration } from "@services/PasswordPolicyConfiguration";
import { cn } from "@utils/Styles";

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

    const resetCapsLockErrors = useCallback(() => {
        setIsCapsLockOnOldPW(false);
        setIsCapsLockOnNewPW(false);
        setIsCapsLockOnRepeatNewPW(false);
    }, []);

    const resetStates = useCallback(() => {
        setOldPassword("");
        setNewPassword("");
        setRepeatNewPassword("");

        resetPasswordErrors();
        resetCapsLockErrors();

        setLoading(false);
    }, [resetPasswordErrors, resetCapsLockErrors]);

    const handleClose = useCallback(() => {
        props.setClosed();
        resetStates();
    }, [props, resetStates]);

    const asyncProcess = useCallback(async () => {
        try {
            setLoading(true);
            const policy = await getPasswordPolicyConfiguration();
            setPPolicy(policy);
            setLoading(false);
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
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) handleClose();
            }}
        >
            <DialogContent className="sm:max-w-xs" showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("Change Password")}</DialogTitle>
                </DialogHeader>
                <fieldset id={"change-password-form"} disabled={loading} className="space-y-1 text-center">
                    <div className="w-full pt-6">
                        <Label htmlFor="old-password" className="sr-only">
                            {translate("Old Password")}
                        </Label>
                        <Input
                            ref={oldPasswordRef}
                            id="old-password"
                            placeholder={translate("Old Password") + " *"}
                            required
                            value={oldPassword}
                            error={oldPasswordError}
                            disabled={disabled}
                            className="w-full"
                            onChange={(v) => setOldPassword(v.target.value)}
                            onFocus={() => setOldPasswordError(false)}
                            type="password"
                            autoCapitalize="off"
                            autoComplete="off"
                            onKeyDown={handleOldPWKeyDown}
                            onKeyUp={useCheckCapsLock(setIsCapsLockOnOldPW)}
                            onBlur={() => setIsCapsLockOnOldPW(false)}
                        />
                        <p
                            className={cn(
                                "text-xs mt-1 h-4",
                                isCapsLockOnOldPW ? "text-destructive" : "text-transparent",
                            )}
                        >
                            {isCapsLockOnOldPW ? translate("Caps Lock is on") : "\u00A0"}
                        </p>
                    </div>
                    <div className="w-full mt-6">
                        <Label htmlFor="new-password" className="sr-only">
                            {translate("New Password")}
                        </Label>
                        <Input
                            ref={newPasswordRef}
                            id="new-password"
                            placeholder={translate("New Password") + " *"}
                            required
                            error={newPasswordError}
                            className="w-full"
                            disabled={disabled}
                            value={newPassword}
                            onChange={(v) => setNewPassword(v.target.value)}
                            onFocus={() => setNewPasswordError(false)}
                            type="password"
                            autoCapitalize="off"
                            autoComplete="off"
                            onKeyDown={handleNewPWKeyDown}
                            onKeyUp={useCheckCapsLock(setIsCapsLockOnNewPW)}
                            onBlur={() => setIsCapsLockOnNewPW(false)}
                        />
                        <p
                            className={cn(
                                "text-xs mt-1 h-4",
                                isCapsLockOnNewPW ? "text-destructive" : "text-transparent",
                            )}
                        >
                            {isCapsLockOnNewPW ? translate("Caps Lock is on") : "\u00A0"}
                        </p>
                        {pPolicy.mode === PasswordPolicyMode.Disabled ? null : (
                            <PasswordMeter value={newPassword} policy={pPolicy} />
                        )}
                    </div>
                    <div className="w-full">
                        <Label htmlFor="repeat-new-password" className="sr-only">
                            {translate("Repeat New Password")}
                        </Label>
                        <Input
                            ref={repeatNewPasswordRef}
                            id="repeat-new-password"
                            placeholder={translate("Repeat New Password") + " *"}
                            required
                            error={repeatNewPasswordError}
                            className="w-full"
                            disabled={disabled}
                            value={repeatNewPassword}
                            onChange={(v) => setRepeatNewPassword(v.target.value)}
                            onFocus={() => setRepeatNewPasswordError(false)}
                            type="password"
                            autoCapitalize="off"
                            autoComplete="off"
                            onKeyDown={handleRepeatNewPWKeyDown}
                            onKeyUp={useCheckCapsLock(setIsCapsLockOnRepeatNewPW)}
                            onBlur={() => setIsCapsLockOnRepeatNewPW(false)}
                        />
                        <p
                            className={cn(
                                "text-xs mt-1 h-4",
                                isCapsLockOnRepeatNewPW ? "text-destructive" : "text-transparent",
                            )}
                        >
                            {isCapsLockOnRepeatNewPW ? translate("Caps Lock is on") : "\u00A0"}
                        </p>
                    </div>
                </fieldset>
                <DialogFooter>
                    <Button id={"password-change-dialog-cancel"} variant={"destructive"} onClick={handleClose}>
                        {translate("Cancel")}
                    </Button>
                    <Button
                        id={"password-change-dialog-submit"}
                        variant={"default"}
                        onClick={handlePasswordChange}
                        disabled={!(oldPassword.length && newPassword.length && repeatNewPassword.length) || loading}
                    >
                        {loading ? <Spinner size={20} /> : null}
                        {translate("Submit")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default ChangePasswordDialog;
