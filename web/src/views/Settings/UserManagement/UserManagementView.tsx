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
import { UserFieldMetadataBody, getUserFieldMetadata } from "@services/UserManagement.js";
import EditUserDialog from "@views/Settings/UserManagement/EditUserDialog";
import NewUserDialog from "@views/Settings/UserManagement/NewUserDialog.tsx";
import VerifyDeleteUserDialog from "@views/Settings/UserManagement/VerifyDeleteUserDialog.tsx";

const UserManagementView = () => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification } = useNotifications();

    const [users, fetchUsers, , fetchUsersError] = useAllUserInfoGET();
    const [selectedUser, setSelectedUser] = useState<null | UserInfo>(null);
    const [userToDelete, setUserToDelete] = useState("");
    const [isEditUserDialogOpen, setIsEditUserDialogOpen] = useState(false);
    const [isNewUserDialogOpen, setIsNewUserDialogOpen] = useState(false);
    const [isVerifyDeleteUserDialogOpen, setIsVerifyDeleteUserDialogOpen] = useState(false);
    const [fieldMetadata, setFieldMetadata] = useState<null | UserFieldMetadataBody>(null);

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
        const fetchFieldMetadata = async () => {
            try {
                const data = await getUserFieldMetadata();
                setFieldMetadata(data);
                console.log(data);
            } catch {
                createErrorNotification(translate("Unable to retrieve field metadata"));
            }
        };

        fetchFieldMetadata();
    }, [createErrorNotification, translate]);

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
                display_name: user.display_name,
                emails: Array.isArray(user.emails) ? user.emails[0] : user.emails,
                has_duo: user.has_duo ? "Yes" : "No",
                has_totp: user.has_totp ? "Yes" : "No",
                has_webauthn: user.has_webauthn ? "Yes" : "No",
                id: index,
                last_logged_in: user.last_logged_in ? new Date(user.last_logged_in).toLocaleString() : "-",
                last_password_change: user.last_password_change
                    ? new Date(user.last_password_change).toLocaleString()
                    : "-",
                method:
                    user.method && (user.has_duo || user.has_totp || user.has_webauthn)
                        ? to2FAString(user.method)
                        : "-",
                user_created_at: user.user_created_at ? new Date(user.user_created_at).toLocaleString() : "-",
                username: user.username,
            };
        });
    }, [users, createErrorNotification]);

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
            headerName: "Actions",
            type: "actions",
            width: 100,
        },
    ];

    return (
        <>
            <EditUserDialog user={selectedUser} open={isEditUserDialogOpen} onClose={handleCloseEditUserDialog} />
            {fieldMetadata && (
                <NewUserDialog open={isNewUserDialogOpen} onClose={handleCloseNewUserDialog} metadata={fieldMetadata} />
            )}
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
                                has_duo: false,
                                has_totp: false,
                                has_webauthn: false,
                                logout_required: false,
                                password_change_required: false,
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
