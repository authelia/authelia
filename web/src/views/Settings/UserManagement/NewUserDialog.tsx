import React, { useEffect, useState } from "react";

import { Autocomplete, Button, Dialog, DialogContent, DialogTitle, FormControl, Grid, TextField } from "@mui/material";
import { useTranslation } from "react-i18next";

import PasswordMeter from "@components/PasswordMeter.tsx";
import { useNotifications } from "@hooks/NotificationsContext";
import { PasswordPolicyConfiguration, PasswordPolicyMode } from "@models/PasswordPolicy.ts";
import { UserInfo, ValidateDisplayName, ValidateEmail, ValidateUsername } from "@models/UserInfo";
import { getPasswordPolicyConfiguration } from "@services/PasswordPolicyConfiguration";
import { postNewUser } from "@services/UserManagement";
import { generateRandomPassword } from "@utils/GeneratePassword.ts";
import VerifyExitDialog from "@views/Settings/Common/VerifyExitDialog";

interface Props {
    open: boolean;
    onClose: () => void;
}
interface NewUser extends UserInfo {
    password: string;
}

const NewUserDialog = (props: Props) => {
    const emptyUser: NewUser = {
        username: "",
        display_name: "",
        emails: [""],
        password: "",
        disabled: false,
        groups: [],
        method: 1,
        has_duo: false,
        has_totp: false,
        has_webauthn: false,
    };

    const passwordLength = 20;
    const { t: translate } = useTranslation("settings");
    const { createSuccessNotification, createErrorNotification } = useNotifications();

    const [newUser, setNewUser] = useState<NewUser>(emptyUser);
    const [changesMade, setChangesMade] = useState(false);
    const [verifyExitDialogOpen, setVerifyExitDialogOpen] = useState(false);
    const [usernameError, setUsernameError] = useState(false);
    const [displayNameError, setDisplayNameError] = useState(false);
    const [emailError, setEmailError] = useState(false);
    const [passwordError, setPasswordError] = useState(false);

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

    useEffect(() => {
        const fetchPasswordPolicy = async () => {
            try {
                const policy = await getPasswordPolicyConfiguration();
                setPPolicy(policy);
            } catch (err) {
                console.error(err);
                createErrorNotification(translate("There was an issue retrieving the password policy"));
            }
        };
        fetchPasswordPolicy().then(() => {});
    }, [setPPolicy, createErrorNotification, translate]);

    useEffect(() => {
        setChangesMade(false);
    }, [setChangesMade]);

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
        setNewUser(emptyUser);
        handleResetErrors();
    };

    const handleResetErrors = () => {
        setUsernameError(false);
        setDisplayNameError(false);
        setEmailError(false);
        setPasswordError(false);
    };

    const userIsEmpty = (user: NewUser): boolean => {
        return (
            user.username.trim() !== "" ||
            user.display_name.trim() !== "" ||
            user.password.trim() !== "" ||
            user.emails[0].trim() !== "" ||
            user.groups.length > 0
        );
    };

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value, type, checked } = event.target;
        setNewUser((prev) => {
            if (!prev) {
                prev = emptyUser;
            }
            let updatedUser: NewUser;
            if (name === "username") {
                updatedUser = {
                    ...prev,
                    username: value,
                };
                setChangesMade(true);
                if (!ValidateUsername(newUser.username)) {
                    setUsernameError(true);
                } else {
                    setUsernameError(false);
                }
            } else if (name === "emails") {
                updatedUser = {
                    ...prev,
                    emails: [value],
                };
                setChangesMade(true);
                if (!ValidateEmail(newUser.emails[0])) {
                    setEmailError(true);
                } else {
                    setEmailError(false);
                }
            } else {
                updatedUser = {
                    ...prev,
                    [name]: type === "checkbox" ? checked : value,
                };
            }
            setChangesMade(userIsEmpty(updatedUser));
            return updatedUser;
        });
    };

    const handleGroupsChange = (_event: React.SyntheticEvent, value: string[]) => {
        setNewUser((prev) => {
            if (!prev) {
                prev = emptyUser;
            }

            const updatedUser = {
                ...prev,
                groups: value,
            };
            setChangesMade(true);
            return updatedUser;
        });
    };

    const handleSave = async () => {
        if (!changesMade) {
            handleSafeClose();
        }
        if (newUser === null) {
            handleSafeClose();
            return;
        }
        handleResetErrors();
        let error = false;

        if (!ValidateUsername(newUser.username)) {
            error = true;
            setUsernameError(true);
        }

        if (!ValidateDisplayName(newUser.display_name)) {
            error = true;
            setDisplayNameError(true);
        }

        if (!ValidateEmail(newUser.emails[0]) && newUser.emails[0] !== "") {
            error = true;
            setEmailError(true);
        }

        if (newUser.password === "") {
            error = true;
            setPasswordError(true);
        }

        if (error) {
            return;
        }

        try {
            await postNewUser(
                newUser.username,
                newUser.display_name,
                newUser.password,
                newUser.disabled ? newUser.disabled : false,
                Array.isArray(newUser.emails) ? newUser.emails[0] : newUser.emails,
                newUser.groups,
            );
            createSuccessNotification(translate("User created successfully."));
            handleClose();
        } catch {
            handleResetErrors();
            createErrorNotification(translate("Error: More informative errors WIP"));
        }
    };

    const handleConfirmExit = () => {
        setVerifyExitDialogOpen(false);
        setNewUser(newUser);
        setChangesMade(false);
        handleSafeClose();
    };

    const handleCancelExit = () => {
        setVerifyExitDialogOpen(false);
    };

    const handleGeneratePassword = () => {
        setNewUser((prev) => {
            if (!prev) {
                return emptyUser;
            }
            return {
                ...prev,
                password: generateRandomPassword(passwordLength),
            };
        });
        setChangesMade(true);
    };

    return (
        <div>
            <Dialog open={props.open} onClose={handleSafeClose} maxWidth="xs" fullWidth>
                <DialogTitle>{translate("New {{item}}", { item: translate("User") })}</DialogTitle>
                <DialogContent>
                    <FormControl variant={"standard"}>
                        <Grid container spacing={1} alignItems={"center"}>
                            <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <TextField
                                    fullWidth
                                    id="enter-username"
                                    label={translate("Username")}
                                    name="username"
                                    error={usernameError}
                                    value={newUser?.username ?? ""}
                                    onChange={handleChange}
                                    required
                                />
                            </Grid>
                            <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <TextField
                                    fullWidth
                                    id="enter-user-display-name"
                                    label={translate("Display Name")}
                                    name="display_name"
                                    error={displayNameError}
                                    value={newUser?.display_name ?? ""}
                                    onChange={handleChange}
                                    required
                                />
                            </Grid>
                            <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <TextField
                                    fullWidth
                                    id="enter-user-password"
                                    label={translate("Password")}
                                    name="password"
                                    error={passwordError}
                                    value={newUser?.password ?? ""}
                                    onChange={handleChange}
                                    required
                                />
                                {pPolicy.mode === PasswordPolicyMode.Disabled ? null : (
                                    <PasswordMeter value={newUser?.password ?? ""} policy={pPolicy} />
                                )}
                                <Button onClick={handleGeneratePassword}>{translate("Generate Password")}</Button>
                            </Grid>
                            <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <TextField
                                    id="enter-user-email"
                                    fullWidth
                                    label={translate("Email")}
                                    name="emails"
                                    error={emailError}
                                    value={Array.isArray(newUser?.emails) ? newUser.emails[0] : (newUser?.emails ?? "")}
                                    onChange={handleChange}
                                />
                            </Grid>
                            <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <Autocomplete
                                    multiple
                                    id="select-user-groups"
                                    options={[]}
                                    value={newUser?.groups || []}
                                    onChange={handleGroupsChange}
                                    freeSolo
                                    renderInput={(params) => (
                                        <TextField {...params} label={translate("Groups")} placeholder="" />
                                    )}
                                />
                            </Grid>
                            <Grid size={{ xs: 12 }} sx={{ pt: 3 }}>
                                <Button
                                    color={"success"}
                                    onClick={handleSave}
                                    disabled={
                                        !changesMade ||
                                        newUser.username.trim() === "" ||
                                        newUser.display_name.trim() === "" ||
                                        newUser.password.trim() === ""
                                    }
                                >
                                    Save
                                </Button>
                                <Button color={"error"} onClick={handleSafeClose}>
                                    Exit
                                </Button>
                            </Grid>
                        </Grid>
                    </FormControl>
                </DialogContent>
            </Dialog>
            <VerifyExitDialog open={verifyExitDialogOpen} onConfirm={handleConfirmExit} onCancel={handleCancelExit} />
        </div>
    );
};

export default NewUserDialog;
