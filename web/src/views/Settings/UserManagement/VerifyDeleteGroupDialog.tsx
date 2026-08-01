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
    const [disableDelete, setDisableDelete] = useState(false);

    const handleGroupNameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const value = event.target.value;
        setGroupNameInput(value);
        setDisableDelete(value === props.groupName);
    };

    const handleDeleteGroup = async () => {
        if (!props.groupName) {
            return;
        }
        try {
            await deleteGroup(props.groupName);
            createSuccessNotification(translate("Group deleted successfully."));
            handleClose();
        } catch (err) {
            console.log(err);
            createErrorNotification(translate("Error deleting group."));
            handleClose();
        }
    };

    const handleClose = () => {
        setGroupNameInput("");
        setDisableDelete(false);
        props.onCancel();
    };

    return (
        <Dialog open={props.open} onClose={handleClose}>
            <DialogTitle>{translate("Delete Group")}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    <Typography>
                        {translate("You are about to delete group {{item}}, enter the group name to continue.", {
                            item: props.groupName,
                        })}
                    </Typography>
                </DialogContentText>
                <TextField
                    fullWidth
                    label={translate("Group Name")}
                    name="groupName"
                    value={groupNameInput}
                    onChange={handleGroupNameChange}
                    required
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={handleClose}>Cancel</Button>
                <Button onClick={handleDeleteGroup} disabled={!disableDelete} color="error">
                    {translate("Delete Group")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default VerifyDeleteGroupDialog;
