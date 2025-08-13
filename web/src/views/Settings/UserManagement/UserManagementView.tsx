import React, { useCallback, useEffect, useMemo, useState } from "react";

import DeleteIcon from "@mui/icons-material/DeleteOutlined";
import EditIcon from "@mui/icons-material/Edit";
import { Button, Stack } from "@mui/material";
import { DataGrid, GridActionsCellItem, GridColDef, GridRowParams, GridRowsProp } from "@mui/x-data-grid";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { useAllUserInfoGET } from "@hooks/UserManagement";
import { UserInfo } from "@models/UserInfo";
import { to2FAString } from "@services/UserInfo";
import EditUserDialog from "@views/Settings/UserManagement/EditUserDialog";
import NewUserDialog from "@views/Settings/UserManagement/NewUserDialog.tsx";
import VerifyDeleteUserDialog from "@views/Settings/UserManagement/VerifyDeleteUserDialog.tsx";

const UserManagementView = () => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification } = useNotifications();

    const [users, fetchUsers, , fetchUsersError] = useAllUserInfoGET();
    const [selectedUser, setSelectedUser] = useState<UserInfo | null>(null);
    const [userToDelete, setUserToDelete] = useState("");
    const [isEditUserDialogOpen, setIsEditUserDialogOpen] = useState(false);
    const [isNewUserDialogOpen, setIsNewUserDialogOpen] = useState(false);
    const [isVerifyDeleteUserDialogOpen, setIsVerifyDeleteUserDialogOpen] = useState(false);

    const handleRowClick = (params: GridRowParams) => {
        if (!users) {
            createErrorNotification(translate("Unable to edit user."));
            return;
        }
        handleOpenEditUserDialog(params.row.username);
    };

    const handleResetState = useCallback(() => {
        setIsEditUserDialogOpen(false);
        setIsVerifyDeleteUserDialogOpen(false);
        setIsNewUserDialogOpen(false);
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
            createErrorNotification("Error fetching User Info");
            return [];
        }

        return users.map((user: UserInfo, index: number) => {
            return {
                id: index,
                username: user.username,
                display_name: user.display_name,
                emails: Array.isArray(user.emails) ? user.emails[0] : user.emails,
                last_logged_in: user.last_logged_in ? new Date(user.last_logged_in).toLocaleString() : "-",
                last_password_change: user.last_password_change
                    ? new Date(user.last_password_change).toLocaleString()
                    : "-",
                user_created_at: user.user_created_at ? new Date(user.user_created_at).toLocaleString() : "-",
                method:
                    user.method && (user.has_duo || user.has_totp || user.has_webauthn)
                        ? to2FAString(user.method)
                        : "-",
                has_webauthn: user.has_webauthn ? "Yes" : "No",
                has_totp: user.has_totp ? "Yes" : "No",
                has_duo: user.has_duo ? "Yes" : "No",
            };
        });
    }, [users, createErrorNotification]);

    const columns: GridColDef[] = [
        { field: "username", headerName: "Username", flex: 1 },
        { field: "display_name", headerName: "Display Name", flex: 1 },
        { field: "emails", headerName: "Email", flex: 1 },
        { field: "last_logged_in", headerName: "Last Log In", flex: 1 },
        { field: "last_password_change", headerName: "Last Password Change", flex: 1 },
        { field: "user_created_at", headerName: "User Created At", flex: 1 },
        { field: "method", headerName: "Default 2FA Method", flex: 1 },
        { field: "has_webauthn", headerName: "WebAuthn?", flex: 1 },
        { field: "has_totp", headerName: "Totp?", flex: 1 },
        { field: "has_duo", headerName: "Duo?", flex: 1 },
        {
            field: "actions",
            type: "actions",
            headerName: "Actions",
            width: 100,
            cellClassName: "actions",
            getActions: (params: GridRowParams) => {
                return [
                    <GridActionsCellItem
                        icon={<EditIcon />}
                        label="Edit"
                        className="textPrimary"
                        onClick={() => handleOpenEditUserDialog(params.row.username)}
                        color="inherit"
                    />,
                    <GridActionsCellItem
                        icon={<DeleteIcon />}
                        label="Delete"
                        onClick={() => handleOpenVerifyDeleteUserDialog(params.row.username)}
                        color="inherit"
                    />,
                ];
            },
        },
    ];

    return (
        <>
            <EditUserDialog user={selectedUser} open={isEditUserDialogOpen} onClose={handleCloseEditUserDialog} />
            <NewUserDialog open={isNewUserDialogOpen} onClose={handleCloseNewUserDialog} />
            <VerifyDeleteUserDialog
                username={userToDelete || ""}
                open={isVerifyDeleteUserDialogOpen}
                onCancel={handleCloseVerifyDeleteUserDialog}
            />
            <Stack direction={"row"} spacing={1} sx={{ mb: 1 }}>
                <Button size="medium" variant="contained" onClick={handleOpenNewUserDialog}>
                    {translate("Add a {{item}}", { item: "user" })}
                </Button>
            </Stack>
            <div style={{ height: 400, width: "100%" }}>
                <DataGrid
                    rows={rows}
                    columns={columns}
                    editMode={"row"}
                    onRowDoubleClick={handleRowClick}
                    checkboxSelection={false}
                    initialState={{
                        columns: {
                            columnVisibilityModel: {
                                password_change_required: false,
                                logout_required: false,
                                has_totp: false,
                                has_webauthn: false,
                                has_duo: false,
                            },
                        },
                        sorting: {
                            sortModel: [{ field: "username", sort: "asc" }],
                        },
                    }}
                />
            </div>
        </>
    );
};

export default UserManagementView;
