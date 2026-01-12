import { useEffect, useState } from "react";

import {
    Autocomplete,
    Box,
    Button,
    Dialog,
    DialogContent,
    DialogTitle,
    FormControl,
    Grid,
    TextField,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { UserInfo } from "@models/UserInfo";
import { ValidateDisplayName, ValidateEmail, ValidateGroup } from "@models/UserManagement.js";
//import { putChangeUser } from "@services/UserManagement";
import VerifyExitDialog from "@views/Settings/Common/VerifyExitDialog";

interface UserChange extends UserInfo {
    password?: string; //this is only used when an admin is changing a user's password.
}

interface Props {
    user: null | UserInfo;
    open: boolean;
    onClose: () => void;
}

const EditUserDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [editedUser, setEditedUser] = useState<null | UserChange>(null);
    const [changesMade, setChangesMade] = useState(false);
    const [verifyExitDialogOpen, setVerifyExitDialogOpen] = useState(false);
    const [displayNameError, setDisplayNameError] = useState(false);
    const [emailError, setEmailError] = useState(false);
    const [groupsError, setGroupsError] = useState(false);

    useEffect(() => {
        if (props.open && props.user) {
            setEditedUser(props.user);
            setChangesMade(false);
        }
    }, [props.open, props.user]);

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = event.target;
        setEditedUser((prev) => {
            if (!prev) return null;
            const editedUser = {
                ...prev,
                [name]: name === "emails" ? [value] : value,
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
            handleClose();
            return;
        }
        if (!editedUser) {
            handleClose();
            return;
        }

        let error = false;
        if (!ValidateDisplayName(editedUser.display_name)) {
            error = true;
            setDisplayNameError(true);
        }

        if (!editedUser.emails?.length || !ValidateEmail(editedUser.emails[0])) {
            error = true;
            setEmailError(true);
        }

        if (editedUser.groups.length > 0) {
            editedUser.groups.forEach((group) => {
                if (!ValidateGroup(group)) {
                    error = true;
                    setGroupsError(true);
                }
            });
        }

        if (error) {
            return;
        }

        try {
            // await putChangeUser(
            //     editedUser.username,
            //     editedUser.display_name,
            //     editedUser.password ? editedUser.password : "",
            //     editedUser.disabled ? editedUser.disabled : false,
            //     editedUser.emails[0],
            //     editedUser.groups,
            // );
            createSuccessNotification(translate("User modified successfully."));
        } catch (err) {
            handleResetErrors();
            console.log(err);
            if ((err as Error).message.includes("Display name")) {
                setDisplayNameError(true);
                createErrorNotification(
                    translate("The supplied {{item}} is not formatted correctly.", { item: "display name" }),
                );
                return;
            }

            if ((err as Error).message.includes("email")) {
                setEmailError(true);
                createErrorNotification(
                    translate("The supplied {{item}} is not formatted correctly.", { item: "email" }),
                );
                return;
            }

            if ((err as Error).message.includes("email")) {
                setGroupsError(true);
                createErrorNotification(
                    translate("The supplied {{item}} is not formatted correctly.", { item: "group" }),
                );
                return;
            }
        }
        handleClose();
    };

    const handleResetErrors = () => {
        setDisplayNameError(false);
        setEmailError(false);
        setGroupsError(false);
    };

    const handleSafeClose = () => {
        if (changesMade) {
            setVerifyExitDialogOpen(true);
        } else {
            handleClose();
        }
    };

    const handleClose = () => {
        handleResetErrors();
        handleResetState();
        props.onClose();
    };

    const handleResetState = () => {
        setEditedUser(props.user);
        setChangesMade(false);
    };

    const handleConfirmExit = () => {
        setVerifyExitDialogOpen(false);
        handleClose();
    };

    const handleCancelExit = () => {
        setVerifyExitDialogOpen(false);
    };

    return (
        <div>
            <Dialog open={props.open} onClose={handleSafeClose} maxWidth="sm" fullWidth>
                <DialogTitle>
                    {translate("Edit {{item}}:", { item: translate("User") })} {props.user?.username}
                </DialogTitle>
                <DialogContent>
                    <FormControl>
                        <Box sx={{ display: "flex", gap: 3 }}>
                            <Grid container spacing={1} alignItems={"center"}>
                                <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                    <TextField
                                        fullWidth
                                        label="Display Name"
                                        name="display_name"
                                        error={displayNameError}
                                        value={editedUser?.display_name ?? ""}
                                        helperText={
                                            displayNameError ? "Only Letters, Numbers, Symbols, and Spaces" : ""
                                        }
                                        onChange={handleChange}
                                    />
                                </Grid>
                                <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                    <TextField
                                        fullWidth
                                        label="Email"
                                        name="emails"
                                        error={emailError}
                                        onChange={handleChange}
                                        helperText={emailError ? "Standard Email Format (john@example.com)" : ""}
                                        value={
                                            Array.isArray(editedUser?.emails)
                                                ? editedUser.emails[0]
                                                : (editedUser?.emails ?? "")
                                        }
                                    />
                                </Grid>
                                <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                    <Autocomplete
                                        multiple
                                        id="select-user-groups"
                                        options={[]}
                                        value={editedUser?.groups || []}
                                        onChange={handleGroupsChange}
                                        freeSolo
                                        renderInput={(params) => (
                                            <TextField
                                                {...params}
                                                error={groupsError}
                                                label="Groups"
                                                placeholder=""
                                                helperText={
                                                    groupsError
                                                        ? "Letters, numbers, and symbols (+-_.,). Up to 100 characters."
                                                        : ""
                                                }
                                            />
                                        )}
                                    />
                                </Grid>
                                <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                    <Button color={"success"} onClick={handleSave} disabled={!changesMade}>
                                        {translate("Save")}
                                    </Button>
                                    <Button color={"error"} onClick={handleSafeClose}>
                                        {translate("Exit")}
                                    </Button>
                                </Grid>
                            </Grid>
                            {/*</Box>*/}
                        </Box>
                    </FormControl>
                </DialogContent>
            </Dialog>
            <VerifyExitDialog open={verifyExitDialogOpen} onConfirm={handleConfirmExit} onCancel={handleCancelExit} />
        </div>
    );
};

export default EditUserDialog;
