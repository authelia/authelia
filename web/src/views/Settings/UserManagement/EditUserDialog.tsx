import React, { useEffect, useState } from "react";

import { Autocomplete, Button, Dialog, DialogContent, DialogTitle, FormControl, Grid2, TextField } from "@mui/material";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { UserInfo, ValidateDisplayName, ValidateEmail, ValidateGroup } from "@models/UserInfo";
import { postChangeUser } from "@services/UserManagement";
import VerifyExitDialog from "@views/Settings/Common/VerifyExitDialog";

interface Props {
    user: UserInfo | null;
    open: boolean;
    onClose: () => void;
}

const EditUserDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const [editedUser, setEditedUser] = useState<UserInfo | null>(null);
    const [changesMade, setChangesMade] = useState(false);
    const [verifyExitDialogOpen, setVerifyExitDialogOpen] = useState(false);
    const [displayNameError, setDisplayNameError] = useState(false);
    const [emailError, setEmailError] = useState(false);
    const [, setGroupsError] = useState(false);

    useEffect(() => {
        setEditedUser(props.user);
        setChangesMade(false);
    }, [props.user]);

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value, type, checked } = event.target;
        setEditedUser((prev) => {
            if (!prev) return null;
            const editedUser = {
                ...prev,
                [name]: type === "checkbox" ? checked : value,
            };
            setChangesMade(JSON.stringify(editedUser) !== JSON.stringify(props.user));
            return editedUser;
        });
    };

    const handleGroupsChange = (_event: React.SyntheticEvent, value: string[]) => {
        setEditedUser((prev) => {
            if (!prev) return null;
            const updatedUser = { ...prev, groups: value };
            setChangesMade(JSON.stringify(updatedUser) !== JSON.stringify(props.user));
            return updatedUser;
        });
    };

    const handleSave = async () => {
        if (!changesMade) {
            return;
        }
        if (editedUser != null) {
            let error = false;
            if (ValidateDisplayName(editedUser.display_name)) {
                error = true;
                setDisplayNameError(true);
            }
            if (ValidateEmail(editedUser.emails[0])) {
                error = true;
                setEmailError(true);
            }
            if (editedUser.groups.length > 0) {
                editedUser.groups.forEach((group) => {
                    if (ValidateGroup(group)) {
                        error = true;
                        setGroupsError(true);
                    }
                });
            }

            if (error) {
                return;
            }
            try {
                await postChangeUser(
                    editedUser.username,
                    editedUser.display_name,
                    Array.isArray(editedUser.emails) ? editedUser.emails[0] : editedUser.emails,
                    editedUser.groups,
                );
                createSuccessNotification(translate("User modified successfully."));
            } catch (err) {
                handleResetErrors();
                console.log(err);
                if ((err as Error).message.includes("")) {
                    createErrorNotification(translate("An error!"));
                }
            }
            handleClose();
        } else {
            handleClose();
            return;
        }
    };

    const handleResetErrors = () => {
        setDisplayNameError(false);
        setEmailError(false);
        setGroupsError(false);
    };

    const handleClose = () => {
        if (changesMade) {
            setVerifyExitDialogOpen(true);
        } else {
            handleResetErrors();
            props.onClose();
        }
    };

    const handleConfirmExit = () => {
        setVerifyExitDialogOpen(false);
        setEditedUser(props.user);
        setChangesMade(false);
        handleClose();
    };

    const handleCancelExit = () => {
        setVerifyExitDialogOpen(false);
    };

    return (
        <div>
            <Dialog open={props.open} onClose={handleClose} maxWidth="xs" fullWidth>
                <DialogTitle>
                    {translate("Edit {{item}}:", { item: translate("User") })} {props.user?.username}
                </DialogTitle>
                <DialogContent>
                    <FormControl>
                        <Grid2 container spacing={1} alignItems={"center"}>
                            <Grid2 size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <TextField
                                    fullWidth
                                    label="Display Name"
                                    name="display_name"
                                    error={displayNameError}
                                    value={editedUser?.display_name ?? ""}
                                    onChange={handleChange}
                                />
                            </Grid2>
                            <Grid2 size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <TextField
                                    fullWidth
                                    label="Email"
                                    name="emails"
                                    error={emailError}
                                    value={
                                        Array.isArray(editedUser?.emails)
                                            ? editedUser.emails[0]
                                            : (editedUser?.emails ?? "")
                                    }
                                    onChange={handleChange}
                                />
                            </Grid2>
                            <Grid2 size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <Autocomplete
                                    multiple
                                    id="select-user-groups"
                                    options={[]}
                                    value={editedUser?.groups || []}
                                    onChange={handleGroupsChange}
                                    freeSolo
                                    renderInput={(params) => <TextField {...params} label="Groups" placeholder="" />}
                                />
                            </Grid2>
                            <Grid2 size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <Button color={"success"} onClick={handleSave} disabled={!changesMade}>
                                    {translate("Save")}
                                </Button>
                                <Button color={"error"} onClick={handleClose}>
                                    {translate("Exit")}
                                </Button>
                            </Grid2>
                        </Grid2>
                    </FormControl>
                </DialogContent>
            </Dialog>
            <VerifyExitDialog open={verifyExitDialogOpen} onConfirm={handleConfirmExit} onCancel={handleCancelExit} />
        </div>
    );
};

export default EditUserDialog;
