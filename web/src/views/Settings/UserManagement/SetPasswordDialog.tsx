import { useState } from "react";

import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    TextField,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext.ts";
import { postChangePasswordForUser } from "@services/UserManagement.ts";

interface Props {
    open: boolean;
    username: string;
    onCancel: () => void;
}

const SetUserPasswordDialog = (props: Props) => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [password, setPassword] = useState<string>("");
    const [confirmPassword, setConfirmPassword] = useState<string>("");
    const [inProgress, setInProgress] = useState(false);

    const passwordsMatch = password !== "" && password === confirmPassword;

    const handlePasswordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setPassword(event.target.value);
    };

    const handleConfirmPasswordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setConfirmPassword(event.target.value);
    };

    const handleSetPassword = async () => {
        if (!props.username || !passwordsMatch) {
            return;
        }

        setInProgress(true);
        try {
            await postChangePasswordForUser(props.username, password);
            createSuccessNotification(translate("Password updated successfully."));
            handleClose();
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Error updating password."));
        } finally {
            setInProgress(false);
        }
    };

    const handleClose = () => {
        setPassword("");
        setConfirmPassword("");
        setInProgress(false);
        props.onCancel();
    };

    return (
        <Dialog open={props.open} onClose={handleClose}>
            <DialogTitle>{translate("Set User Password")}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    <Typography>
                        {translate("Set a new password for user {{item}}.", {
                            item: props.username,
                        })}
                    </Typography>
                </DialogContentText>
                <TextField
                    fullWidth
                    label={translate("New Password")}
                    name="password"
                    type="password"
                    value={password}
                    onChange={handlePasswordChange}
                    required
                    margin="dense"
                    disabled={inProgress}
                />
                <TextField
                    fullWidth
                    label={translate("Confirm Password")}
                    name="confirmPassword"
                    type="password"
                    value={confirmPassword}
                    onChange={handleConfirmPasswordChange}
                    required
                    margin="dense"
                    error={confirmPassword !== "" && !passwordsMatch}
                    helperText={confirmPassword !== "" && !passwordsMatch ? translate("Passwords do not match") : ""}
                    disabled={inProgress}
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={handleClose} disabled={inProgress}>
                    {translate("Cancel")}
                </Button>
                <Button onClick={handleSetPassword} disabled={!passwordsMatch || inProgress} color="primary">
                    {translate("Set Password")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default SetUserPasswordDialog;
