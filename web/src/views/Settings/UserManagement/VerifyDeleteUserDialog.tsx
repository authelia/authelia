import React, { useState } from "react";

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
import { deleteDeleteUser } from "@services/UserManagement.ts";

interface Props {
    open: boolean;
    username: string;
    onCancel: () => void;
}

const VerifyDeleteUserDialog = (props: Props) => {
    const { t: translate } = useTranslation("settings");
    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const [usernameInput, setUsernameInput] = useState<string>("");
    const [disableDelete, setDisableDelete] = useState(false);

    const handleUsernameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const value = event.target.value;
        setUsernameInput(value);
        setDisableDelete(value === props.username);
    };

    const handleDeleteUser = async () => {
        if (!props.username) {
            return;
        }
        try {
            await deleteDeleteUser(props.username);
            createSuccessNotification(translate("User deleted successfully."));
            handleClose();
        } catch (err) {
            console.log(err);
            createErrorNotification(translate("Error deleting user."));
            handleClose();
        }
    };

    const handleClose = () => {
        setUsernameInput("");
        setDisableDelete(false);
        props.onCancel();
    };

    return (
        <Dialog open={props.open} onClose={handleClose}>
            <DialogTitle>{translate("Delete User")}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    <Typography>
                        {translate("You are about to delete user {{item}}, enter their username to continue.", {
                            item: props.username,
                        })}
                    </Typography>
                </DialogContentText>
                <TextField
                    fullWidth
                    label={translate("Username")}
                    name="username"
                    value={usernameInput}
                    onChange={handleUsernameChange}
                    required
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={handleClose}>Cancel</Button>
                <Button onClick={handleDeleteUser} disabled={!disableDelete} color="error">
                    {translate("Delete User")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default VerifyDeleteUserDialog;
