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
import { deleteGroup } from "@services/GroupManagement";

interface Props {
    open: boolean;
    groupName: string;
    onCancel: () => void;
}

const VerifyDeleteGroupDialog = (props: Props) => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [groupNameInput, setGroupNameInput] = useState<string>("");

    const canDelete = groupNameInput === props.groupName;

    const handleGroupNameChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
        setGroupNameInput(event.target.value);
    }, []);

    const handleClose = useCallback(() => {
        setGroupNameInput("");
        props.onCancel();
    }, [props]);

    const handleDeleteGroup = useCallback(async () => {
        if (!props.groupName || !canDelete) {
            return;
        }

        try {
            await deleteGroup(props.groupName);
            createSuccessNotification(translate("Group deleted successfully."));
            handleClose();
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Error deleting group."));
        }
    }, [canDelete, createErrorNotification, createSuccessNotification, handleClose, props.groupName, translate]);

    return (
        <Dialog
            open={props.open}
            onOpenChange={(open) => {
                if (!open) handleClose();
            }}
        >
            <DialogContent id="verify-delete-group-dialog" className="sm:max-w-md" showCloseButton={false}>
                <DialogHeader>
                    <DialogTitle>{translate("Delete Group")}</DialogTitle>
                    <DialogDescription>
                        {translate("You are about to delete group {{item}}, enter the group name to continue.", {
                            item: props.groupName,
                        })}
                    </DialogDescription>
                </DialogHeader>
                <div className="space-y-2">
                    <Label htmlFor="verify-delete-group-input">{translate("Group Name")}</Label>
                    <Input
                        id="verify-delete-group-input"
                        name="groupName"
                        required
                        value={groupNameInput}
                        onChange={handleGroupNameChange}
                    />
                </div>
                <DialogFooter>
                    <Button id="verify-delete-group-cancel" variant="ghost" onClick={handleClose}>
                        {translate("Cancel")}
                    </Button>
                    <Button
                        id="verify-delete-group-confirm"
                        color="destructive"
                        disabled={!canDelete}
                        onClick={handleDeleteGroup}
                    >
                        {translate("Delete Group")}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default VerifyDeleteGroupDialog;
