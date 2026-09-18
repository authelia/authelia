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
import { Input } from "@components/UI/Input";
import { Label } from "@components/UI/Label";
import { useNotifications } from "@contexts/NotificationsContext";
import { deleteUser } from "@services/UserManagement";

interface Props {
    open: boolean;
    username: string;
    onCancel: () => void;
}

const VerifyDeleteUserDialog = (props: Props) => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [usernameInput, setUsernameInput] = useState<string>("");

    const canDelete = usernameInput === props.username;

    const handleUsernameChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
        setUsernameInput(event.target.value);
    }, []);

    const handleClose = useCallback(() => {
        setUsernameInput("");
        props.onCancel();
    }, [props]);

    const handleDeleteUser = useCallback(async () => {
        if (!props.username || !canDelete) {
            return;
        }

        try {
            await deleteUser(props.username);
            createSuccessNotification(translate("User deleted successfully."));
            handleClose();
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Error deleting user."));
        }
    }, [canDelete, createErrorNotification, createSuccessNotification, handleClose, props.username, translate]);

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) handleClose();
            }}
        >
            <DialogContent id="verify-delete-user-dialog" className="sm:max-w-md" showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("Delete User")}</DialogTitle>
                    <DialogDescription>
                        {translate("You are about to delete user {{item}}, enter their username to continue.", {
                            item: props.username,
                        })}
                    </DialogDescription>
                </DialogHeader>
                <div className="space-y-2">
                    <Label htmlFor="verify-delete-user-input">{translate("Username")}</Label>
                    <Input
                        id="verify-delete-user-input"
                        name="username"
                        required
                        value={usernameInput}
                        onChange={handleUsernameChange}
                    />
                </div>
                <DialogFooter>
                    <Button id="verify-delete-user-cancel" variant="ghost" onClick={handleClose}>
                        {translate("Cancel")}
                    </Button>
                    <Button
                        id="verify-delete-user-confirm"
                        color="destructive"
                        disabled={!canDelete}
                        onClick={handleDeleteUser}
                    >
                        {translate("Delete User")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default VerifyDeleteUserDialog;
