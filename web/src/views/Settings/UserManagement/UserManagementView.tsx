import { useCallback, useEffect, useMemo, useState } from "react";

import { ForwardToInbox, LockReset, MoreVert } from "@mui/icons-material";
import DeleteIcon from "@mui/icons-material/DeleteOutlined";
import EditIcon from "@mui/icons-material/Edit";
import {
    Box,
    Button,
    Divider,
    ListItemIcon,
    ListItemText,
    Menu,
    MenuItem,
    Stack,
    Typography,
    useTheme,
} from "@mui/material";
import { DataGrid, GridActionsCellItem, GridColDef, GridRowParams, GridRowsProp } from "@mui/x-data-grid";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@contexts/NotificationsContext";
import { useAllUserInfoGET } from "@hooks/UserManagement";
import { UserDetailsExtended } from "@models/UserManagement.ts";
import { Method2FA, to2FAString, toSecondFactorMethod } from "@services/UserInfo";
import { postSendResetPasswordEmailForUser } from "@services/UserManagement.ts";
import VerifyActionDialog from "@views/Settings/Common/VerifyActionDialog.tsx";
import EditUserDialog from "@views/Settings/UserManagement/EditUserDialog";
import NewUserDialog from "@views/Settings/UserManagement/NewUserDialog.tsx";
import SetUserPasswordDialog from "@views/Settings/UserManagement/SetPasswordDialog.tsx";
import VerifyDeleteUserDialog from "@views/Settings/UserManagement/VerifyDeleteUserDialog.tsx";

const UserManagementView = () => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();
    const theme = useTheme();

    const [users, fetchUsers, , fetchUsersError] = useAllUserInfoGET();
    const [selectedUser, setSelectedUser] = useState<null | UserDetailsExtended>(null);
    const [userToDelete, setUserToDelete] = useState("");
    const [userForPasswordReset, setUserForPasswordReset] = useState("");
    const [userForPasswordChange, setUserForPasswordChange] = useState("");

    const [isEditUserDialogOpen, setIsEditUserDialogOpen] = useState(false);
    const [isNewUserDialogOpen, setIsNewUserDialogOpen] = useState(false);
    const [isVerifyDeleteUserDialogOpen, setIsVerifyDeleteUserDialogOpen] = useState(false);
    const [isPasswordResetDialogOpen, setIsPasswordResetDialogOpen] = useState(false);
    const [isPasswordChangeDialogOpen, setIsPasswordChangeDialogOpen] = useState(false);

    const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
    const [menuUsername, setMenuUsername] = useState<string>("");
    const isMenuOpen = Boolean(anchorEl);

    const handleRowClick = (params: GridRowParams) => {
        if (!users) {
            createErrorNotification(translate("Unable to edit user"));
            return;
        }
        handleOpenEditUserDialog(params.row.username);
    };

    const handleResetState = useCallback(() => {
        setIsEditUserDialogOpen(false);
        setIsVerifyDeleteUserDialogOpen(false);
        setIsNewUserDialogOpen(false);
        setIsPasswordResetDialogOpen(false);
        setIsPasswordChangeDialogOpen(false);
    }, []);

    const handleOpenEditUserDialog = useCallback(
        (username: string) => {
            const user = users?.find((user) => user.username === username);
            if (!user) {
                return;
            }
            setSelectedUser(user);
            handleResetState();
            setIsEditUserDialogOpen(true);
        },
        [users, handleResetState, setIsEditUserDialogOpen],
    );

    const handleCloseEditUserDialog = useCallback(() => {
        setIsEditUserDialogOpen(false);
        handleResetState();
        fetchUsers();
    }, [handleResetState, setIsEditUserDialogOpen, fetchUsers]);

    const handleOpenNewUserDialog = useCallback(() => {
        setIsNewUserDialogOpen(true);
    }, [setIsNewUserDialogOpen]);

    const handleCloseNewUserDialog = useCallback(() => {
        setIsNewUserDialogOpen(false);
        handleResetState();
        fetchUsers();
    }, [handleResetState, setIsNewUserDialogOpen, fetchUsers]);

    const handleOpenVerifyDeleteUserDialog = useCallback(
        (username: string) => {
            setUserToDelete(username);
            setIsVerifyDeleteUserDialogOpen(true);
        },
        [setIsVerifyDeleteUserDialogOpen],
    );

    const handleCloseVerifyDeleteUserDialog = useCallback(() => {
        setIsVerifyDeleteUserDialogOpen(false);
        handleResetState();
        fetchUsers();
    }, [handleResetState, setIsVerifyDeleteUserDialogOpen, fetchUsers]);

    const handleOpenPasswordResetDialog = useCallback((username: string) => {
        setUserForPasswordReset(username);
        setIsPasswordResetDialogOpen(true);
    }, []);

    const handleClosePasswordResetDialog = useCallback(() => {
        setIsPasswordResetDialogOpen(false);
        handleResetState();
        fetchUsers();
    }, [handleResetState, fetchUsers]);

    const handleSendPasswordResetEmail = async (username: string) => {
        try {
            await postSendResetPasswordEmailForUser(username);
            createSuccessNotification(translate("Password reset email sent successfully"));
            handleClosePasswordResetDialog();
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Error sending password reset email"));
        }
    };

    const handleOpenPasswordChangeDialog = useCallback((username: string) => {
        setUserForPasswordChange(username);
        setIsPasswordChangeDialogOpen(true);
    }, []);

    const handleClosePasswordChangeDialog = useCallback(() => {
        setIsPasswordChangeDialogOpen(false);
        handleResetState();
        fetchUsers();
    }, [handleResetState, fetchUsers]);

    const handleOpenActionMenu = useCallback((event: React.MouseEvent<HTMLElement>, username: string) => {
        event.stopPropagation();
        setAnchorEl(event.currentTarget);
        setMenuUsername(username);
    }, []);

    const handleCloseMenu = () => {
        setAnchorEl(null);
        setMenuUsername("");
    };

    type MenuAction = "change-password" | "delete-user" | "edit-user" | "send-reset-email";
    const handleMenuAction = useCallback(
        (action: MenuAction) => {
            switch (action) {
                case "send-reset-email":
                    handleOpenPasswordResetDialog(menuUsername);
                    break;
                case "change-password":
                    handleOpenPasswordChangeDialog(menuUsername);
                    break;
                case "edit-user":
                    handleOpenEditUserDialog(menuUsername);
                    break;
                case "delete-user":
                    handleOpenVerifyDeleteUserDialog(menuUsername);
                    break;
                default:
                    console.log(`Unhandled action: ${action} for user: ${menuUsername}`);
            }
            handleCloseMenu();
        },
        [
            menuUsername,
            handleOpenPasswordResetDialog,
            handleOpenPasswordChangeDialog,
            handleOpenEditUserDialog,
            handleOpenVerifyDeleteUserDialog,
        ],
    );

    useEffect(() => {
        fetchUsers();
    }, [fetchUsers]);

    useEffect(() => {
        if (fetchUsersError) {
            createErrorNotification(translate("There was an issue retrieving user info"));
        }
    }, [fetchUsersError, createErrorNotification, translate]);

    const rows: GridRowsProp = useMemo(() => {
        if (!users) {
            return [];
        }

        if (!Array.isArray(users)) {
            createErrorNotification(translate("There was an issue retrieving user info"));
            return [];
        }

        return users.map((user: UserDetailsExtended, index: number) => {
            const methodEnum = user.method ? toSecondFactorMethod(user.method as Method2FA) : undefined;

            return {
                display_name: user.display_name,
                emails: user.mail,
                has_duo: user.has_duo ? "Yes" : "No",
                has_totp: user.has_totp ? "Yes" : "No",
                has_webauthn: user.has_webauthn ? "Yes" : "No",
                id: index,
                last_logged_in: user.last_logged_in ? new Date(user.last_logged_in).toLocaleString() : "-",
                last_password_change: user.last_password_change
                    ? new Date(user.last_password_change).toLocaleString()
                    : "-",
                method:
                    methodEnum && (user.has_duo || user.has_totp || user.has_webauthn) ? to2FAString(methodEnum) : "-",
                user_created_at: user.user_created_at ? new Date(user.user_created_at).toLocaleString() : "-",
                username: user.username,
            };
        });
    }, [users, createErrorNotification, translate]);

    const columns: GridColDef[] = [
        { field: "username", flex: 1, headerName: "Username" },
        { field: "display_name", flex: 1, headerName: "Display Name" },
        { field: "emails", flex: 1, headerName: "Email" },
        { field: "last_logged_in", flex: 1, headerName: "Last Log In" },
        { field: "last_password_change", flex: 1, headerName: "Last Password Change" },
        { field: "user_created_at", flex: 1, headerName: "User Created At" },
        { field: "method", flex: 1, headerName: "Default 2FA Method" },
        { field: "has_webauthn", flex: 1, headerName: "WebAuthn?" },
        { field: "has_totp", flex: 1, headerName: "Totp?" },
        { field: "has_duo", flex: 1, headerName: "Duo?" },
        {
            cellClassName: "actions",
            field: "actions",
            getActions: (params: GridRowParams) => {
                return [
                    <GridActionsCellItem
                        key="edit"
                        icon={<EditIcon />}
                        label={translate("Edit this {{item}}", { item: "user" })}
                        className="textPrimary"
                        onClick={() => handleOpenEditUserDialog(params.row.username)}
                        color="inherit"
                    />,
                    <GridActionsCellItem
                        key="delete"
                        icon={<DeleteIcon />}
                        label={translate("Delete {{item}}", { item: "user" })}
                        onClick={() => handleOpenVerifyDeleteUserDialog(params.row.username)}
                        color="inherit"
                    />,
                    <GridActionsCellItem
                        key="more"
                        icon={<MoreVert />}
                        label={translate("More actions")}
                        onClick={(e) => handleOpenActionMenu(e as React.MouseEvent<HTMLElement>, params.row.username)}
                        color="inherit"
                    />,
                ];
            },
            headerName: "Actions",
            type: "actions",
            width: 100,
        },
    ];

    return (
        <Box
            sx={{
                display: "flex",
                flexDirection: "column",
                height: "85vh",
                minHeight: 0,
                width: "100%",
            }}
        >
            <Typography variant="h4" sx={{ mb: 2 }}>
                {translate("User Management")}
            </Typography>

            <EditUserDialog
                key={selectedUser?.username || "new"}
                user={selectedUser}
                open={isEditUserDialogOpen}
                onClose={handleCloseEditUserDialog}
            />
            <NewUserDialog open={isNewUserDialogOpen} onClose={handleCloseNewUserDialog} />
            <VerifyDeleteUserDialog
                username={userToDelete || ""}
                open={isVerifyDeleteUserDialogOpen}
                onCancel={handleCloseVerifyDeleteUserDialog}
            />

            <SetUserPasswordDialog
                open={isPasswordChangeDialogOpen}
                username={userForPasswordChange}
                onCancel={handleClosePasswordChangeDialog}
            />

            <VerifyActionDialog
                open={isPasswordResetDialogOpen}
                title={translate("Reset Password")}
                message={
                    translate(
                        "You are about to send a password reset email to {{user}} {{username}}, would you like to continue",
                        { user: "user", username: userForPasswordReset },
                    ) + "?"
                }
                cancelText={translate("Cancel")}
                confirmText={translate("Send Password Reset Email")}
                onConfirm={() => {
                    handleSendPasswordResetEmail(userForPasswordReset);
                }}
                onCancel={handleClosePasswordResetDialog}
            />

            <Menu
                anchorEl={anchorEl}
                open={isMenuOpen}
                onClose={handleCloseMenu}
                anchorOrigin={{
                    horizontal: "right",
                    vertical: "bottom",
                }}
                transformOrigin={{
                    horizontal: "right",
                    vertical: "top",
                }}
            >
                <MenuItem onClick={() => handleMenuAction("send-reset-email")}>
                    <ListItemIcon>
                        <ForwardToInbox fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>Send Password Reset Email</ListItemText>
                </MenuItem>
                <MenuItem onClick={() => handleMenuAction("change-password")}>
                    <ListItemIcon>
                        <LockReset fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>Change Password</ListItemText>
                </MenuItem>

                <Divider />

                <MenuItem onClick={() => handleMenuAction("edit-user")}>
                    <ListItemIcon>
                        <EditIcon fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>Edit User</ListItemText>
                </MenuItem>
                <MenuItem onClick={() => handleMenuAction("delete-user")}>
                    <ListItemIcon>
                        <DeleteIcon sx={{ color: theme.palette.error.main }} fontSize="small" />
                    </ListItemIcon>
                    <ListItemText sx={{ color: theme.palette.error.main }}>Delete User</ListItemText>
                </MenuItem>
            </Menu>

            <Stack direction={"row"} spacing={1} sx={{ mb: 1 }}>
                <Button size="medium" variant="contained" onClick={handleOpenNewUserDialog}>
                    {translate("Add a {{item}}", { item: "user" })}
                </Button>
            </Stack>
            <Box style={{ flex: 1, minHeight: 0, minWidth: 0, width: "100%" }}>
                <DataGrid
                    rows={rows}
                    columns={columns}
                    editMode={"row"}
                    onRowDoubleClick={handleRowClick}
                    checkboxSelection={false}
                    sx={{
                        height: "100%",
                        width: "100%",
                    }}
                    initialState={{
                        columns: {
                            columnVisibilityModel: {
                                has_duo: false,
                                has_totp: false,
                                has_webauthn: false,
                                logout_required: false,
                                password_change_required: false,
                            },
                        },
                        pagination: {
                            paginationModel: { page: 0, pageSize: 25 },
                        },
                        sorting: {
                            sortModel: [{ field: "username", sort: "asc" }],
                        },
                    }}
                />
            </Box>
        </Box>
    );
};

export default UserManagementView;
