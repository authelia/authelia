import { type ChangeEvent, useCallback, useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/UI/Dialog";
import { FieldError } from "@components/UI/Field";
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { useNotifications } from "@contexts/NotificationsContext";
import { postChangePasswordForUser } from "@services/UserManagement";

interface Props {
    open: boolean;
    username: string;
    onCancel: () => void;
}

const SetPasswordDialog = (props: Props) => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [password, setPassword] = useState<string>("");
    const [confirmPassword, setConfirmPassword] = useState<string>("");
    const [submitting, setSubmitting] = useState(false);

    const passwordsMatch = password !== "" && password === confirmPassword;
    const showMismatch = confirmPassword !== "" && !passwordsMatch;

    const handlePasswordChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
        setPassword(event.target.value);
    }, []);

    const handleConfirmPasswordChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
        setConfirmPassword(event.target.value);
    }, []);

    const handleClose = useCallback(() => {
        setPassword("");
        setConfirmPassword("");
        setSubmitting(false);
        props.onCancel();
    }, [props]);

    const handleSetPassword = useCallback(async () => {
        if (!props.username || !passwordsMatch) {
            return;
        }

        setSubmitting(true);

        try {
            await postChangePasswordForUser(props.username, password);
            createSuccessNotification(translate("Password updated successfully."));
            handleClose();
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Error updating password."));
            setSubmitting(false);
        }
    }, [
        createErrorNotification,
        createSuccessNotification,
        handleClose,
        password,
        passwordsMatch,
        props.username,
        translate,
    ]);

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open && !submitting) handleClose();
            }}
        >
            <DialogContent id="set-password-dialog" className="sm:max-w-md" showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("Set User Password")}</DialogTitle>
                    <DialogDescription>
                        {translate("Set a new password for user {{item}}.", { item: props.username })}
                    </DialogDescription>
                </DialogHeader>
                <div className="space-y-2">
                    <Label htmlFor="set-password-password">{translate("New Password")}</Label>
                    <Input
                        id="set-password-password"
                        name="password"
                        type="password"
                        required
                        value={password}
                        disabled={submitting}
                        onChange={handlePasswordChange}
                    />
                </div>
                <div className="space-y-2">
                    <Label htmlFor="set-password-confirm">{translate("Confirm Password")}</Label>
                    <Input
                        id="set-password-confirm"
                        name="confirmPassword"
                        type="password"
                        required
                        error={showMismatch}
                        value={confirmPassword}
                        disabled={submitting}
                        onChange={handleConfirmPasswordChange}
                    />
                    {showMismatch ? <FieldError>{translate("Passwords do not match")}</FieldError> : null}
                </div>
                <DialogFooter>
                    <Button id="set-password-cancel" variant="ghost" disabled={submitting} onClick={handleClose}>
                        {translate("Cancel")}
                    </Button>
                    <Button
                        id="set-password-submit"
                        color="primary"
                        disabled={!passwordsMatch || submitting}
                        onClick={handleSetPassword}
                    >
                        {translate("Set Password")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default SetPasswordDialog;
