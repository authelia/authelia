import { useCallback, useEffect, useMemo, useState } from "react";

import { KeyRound, Mail, MoreVertical, Pencil, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";

import { ColumnDef, DataTable, PaginationState, RowAction, SortState } from "@components/DataTable";
import { Button } from "@components/UI/Button";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
} from "@components/UI/DropdownMenu";
import { useNotifications } from "@contexts/NotificationsContext";
import { useAllUserInfoGET } from "@hooks/UserManagement";
import { UserDetailsExtended } from "@models/UserManagement";
import { Method2FA, to2FAString, toSecondFactorMethodOptional } from "@services/UserInfo";
import { postSendResetPasswordEmailForUser } from "@services/UserManagement";
import VerifyActionDialog from "@views/Settings/Common/VerifyActionDialog";
import EditUserDialog from "@views/Settings/UserManagement/EditUserDialog";
import NewUserDialog from "@views/Settings/UserManagement/NewUserDialog";
import SetPasswordDialog from "@views/Settings/UserManagement/SetPasswordDialog";
import VerifyDeleteUserDialog from "@views/Settings/UserManagement/VerifyDeleteUserDialog";

interface MoreMenuAnchor {
    username: string;
    element: HTMLElement;
}

const UserManagementView = () => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [users, fetchUsers, loading, fetchUsersError] = useAllUserInfoGET();

    const [sort, setSort] = useState<SortState>({ direction: "asc", field: "username" });
    const [pagination, setPagination] = useState<Omit<PaginationState, "total">>({ page: 1, pageSize: 25 });
    const [filter, setFilter] = useState("");

    const [selectedUser, setSelectedUser] = useState<null | UserDetailsExtended>(null);
    const [userToDelete, setUserToDelete] = useState("");
    const [userForPasswordReset, setUserForPasswordReset] = useState("");
    const [userForPasswordChange, setUserForPasswordChange] = useState("");

    const [isEditUserDialogOpen, setIsEditUserDialogOpen] = useState(false);
    const [isNewUserDialogOpen, setIsNewUserDialogOpen] = useState(false);
    const [isVerifyDeleteUserDialogOpen, setIsVerifyDeleteUserDialogOpen] = useState(false);
    const [isPasswordResetDialogOpen, setIsPasswordResetDialogOpen] = useState(false);
    const [isPasswordChangeDialogOpen, setIsPasswordChangeDialogOpen] = useState(false);

    const [moreMenuAnchor, setMoreMenuAnchor] = useState<MoreMenuAnchor | null>(null);

    useEffect(() => {
        fetchUsers();
    }, [fetchUsers]);

    useEffect(() => {
        if (fetchUsersError) {
            createErrorNotification(translate("There was an issue retrieving user info"));
        }
    }, [createErrorNotification, fetchUsersError, translate]);

    const handleOpenEditUserDialog = useCallback(
        (username: string) => {
            const user = users?.find((candidate) => candidate.username === username);

            if (!user) {
                return;
            }

            setSelectedUser(user);
            setIsEditUserDialogOpen(true);
        },
        [users],
    );

    const handleCloseEditUserDialog = useCallback(() => {
        setIsEditUserDialogOpen(false);
        fetchUsers();
    }, [fetchUsers]);

    const handleOpenNewUserDialog = useCallback(() => {
        setIsNewUserDialogOpen(true);
    }, []);

    const handleCloseNewUserDialog = useCallback(() => {
        setIsNewUserDialogOpen(false);
        fetchUsers();
    }, [fetchUsers]);

    const handleOpenVerifyDeleteUserDialog = useCallback((username: string) => {
        setUserToDelete(username);
        setIsVerifyDeleteUserDialogOpen(true);
    }, []);

    const handleCloseVerifyDeleteUserDialog = useCallback(() => {
        setIsVerifyDeleteUserDialogOpen(false);
        fetchUsers();
    }, [fetchUsers]);

    const handleOpenPasswordResetDialog = useCallback((username: string) => {
        setUserForPasswordReset(username);
        setIsPasswordResetDialogOpen(true);
    }, []);

    const handleClosePasswordResetDialog = useCallback(() => {
        setIsPasswordResetDialogOpen(false);
        fetchUsers();
    }, [fetchUsers]);

    const handleOpenPasswordChangeDialog = useCallback((username: string) => {
        setUserForPasswordChange(username);
        setIsPasswordChangeDialogOpen(true);
    }, []);

    const handleClosePasswordChangeDialog = useCallback(() => {
        setIsPasswordChangeDialogOpen(false);
        fetchUsers();
    }, [fetchUsers]);

    const handleSendPasswordResetEmail = useCallback(async () => {
        try {
            await postSendResetPasswordEmailForUser(userForPasswordReset);
            createSuccessNotification(translate("Password reset email sent successfully"));
            handleClosePasswordResetDialog();
        } catch (err) {
            console.error(err);
            createErrorNotification(translate("Error sending password reset email"));
        }
    }, [
        createErrorNotification,
        createSuccessNotification,
        handleClosePasswordResetDialog,
        translate,
        userForPasswordReset,
    ]);

    const handleCloseMoreMenu = useCallback(() => {
        setMoreMenuAnchor(null);
    }, []);

    const handleOpenMoreMenu = useCallback((username: string) => {
        const element = document.getElementById(`user-row-${username}-more`);

        if (!element) {
            return;
        }

        setMoreMenuAnchor({ element, username });
    }, []);

    const handleMenuResetEmail = useCallback(() => {
        const username = moreMenuAnchor?.username ?? "";

        handleCloseMoreMenu();
        handleOpenPasswordResetDialog(username);
    }, [handleCloseMoreMenu, handleOpenPasswordResetDialog, moreMenuAnchor]);

    const handleMenuChangePassword = useCallback(() => {
        const username = moreMenuAnchor?.username ?? "";

        handleCloseMoreMenu();
        handleOpenPasswordChangeDialog(username);
    }, [handleCloseMoreMenu, handleOpenPasswordChangeDialog, moreMenuAnchor]);

    const handleMenuEdit = useCallback(() => {
        const username = moreMenuAnchor?.username ?? "";

        handleCloseMoreMenu();
        handleOpenEditUserDialog(username);
    }, [handleCloseMoreMenu, handleOpenEditUserDialog, moreMenuAnchor]);

    const handleMenuDelete = useCallback(() => {
        const username = moreMenuAnchor?.username ?? "";

        handleCloseMoreMenu();
        handleOpenVerifyDeleteUserDialog(username);
    }, [handleCloseMoreMenu, handleOpenVerifyDeleteUserDialog, moreMenuAnchor]);

    const columns = useMemo<ColumnDef<UserDetailsExtended>[]>(
        () => [
            { field: "username", header: translate("Username"), value: (row) => row.username },
            {
                field: "display_name",
                header: translate("Display Name"),
                value: (row) => row.display_name || "-",
            },
            {
                field: "mail",
                header: translate("Email"),
                value: (row) => (row.mail && row.mail.length > 0 ? row.mail.join(", ") : "-"),
            },
            {
                field: "last_logged_in",
                header: translate("Last Log In"),
                value: (row) => (row.last_logged_in ? new Date(row.last_logged_in).toLocaleString() : "-"),
            },
            {
                field: "last_password_change",
                header: translate("Last Password Change"),
                value: (row) => (row.last_password_change ? new Date(row.last_password_change).toLocaleString() : "-"),
            },
            {
                field: "user_created_at",
                header: translate("User Created At"),
                value: (row) => (row.user_created_at ? new Date(row.user_created_at).toLocaleString() : "-"),
            },
            {
                field: "method",
                header: translate("Default 2FA Method"),
                value: (row) => {
                    if (!row.has_totp && !row.has_webauthn && !row.has_duo) {
                        return "-";
                    }

                    const method = toSecondFactorMethodOptional(row.method as Method2FA | undefined);

                    return method ? to2FAString(method) : "-";
                },
            },
            {
                field: "has_webauthn",
                header: translate("WebAuthn?"),
                hidden: true,
                value: (row) => (row.has_webauthn ? translate("Yes") : translate("No")),
            },
            {
                field: "has_totp",
                header: translate("Totp?"),
                hidden: true,
                value: (row) => (row.has_totp ? translate("Yes") : translate("No")),
            },
            {
                field: "has_duo",
                header: translate("Duo?"),
                hidden: true,
                value: (row) => (row.has_duo ? translate("Yes") : translate("No")),
            },
        ],
        [translate],
    );

    const rowActions = useMemo<RowAction<UserDetailsExtended>[]>(
        () => [
            {
                icon: <Pencil />,
                id: () => "edit",
                label: translate("Edit this {{item}}", { item: translate("User") }),
                onClick: (row) => handleOpenEditUserDialog(row.username),
            },
            {
                icon: <Trash2 />,
                id: () => "delete",
                label: translate("Delete this {{item}}", { item: translate("User") }),
                onClick: (row) => handleOpenVerifyDeleteUserDialog(row.username),
            },
            {
                icon: <MoreVertical />,
                id: () => "more",
                label: translate("More actions"),
                onClick: (row) => handleOpenMoreMenu(row.username),
            },
        ],
        [handleOpenEditUserDialog, handleOpenMoreMenu, handleOpenVerifyDeleteUserDialog, translate],
    );

    const filteredSortedUsers = useMemo(() => {
        const all = users ?? [];

        const matches = (row: UserDetailsExtended) =>
            !filter || columns.some((column) => column.value(row).toLowerCase().includes(filter.toLowerCase()));

        const filtered = all.filter(matches);
        const column = columns.find((candidate) => candidate.field === sort.field);

        const sorted = column
            ? [...filtered].sort((a, b) => {
                  const result = column.value(a).localeCompare(column.value(b));

                  return sort.direction === "asc" ? result : -result;
              })
            : filtered;

        return sorted;
    }, [columns, filter, sort, users]);

    const rows = useMemo(
        () =>
            filteredSortedUsers.slice(
                (pagination.page - 1) * pagination.pageSize,
                pagination.page * pagination.pageSize,
            ),
        [filteredSortedUsers, pagination.page, pagination.pageSize],
    );

    return (
        <div className="flex flex-col">
            <h4 className="mb-4 text-2xl font-semibold">{translate("User Management")}</h4>

            <EditUserDialog
                key={selectedUser?.username || "new"}
                onClose={handleCloseEditUserDialog}
                open={isEditUserDialogOpen}
                user={selectedUser}
            />
            <NewUserDialog onClose={handleCloseNewUserDialog} open={isNewUserDialogOpen} />
            <VerifyDeleteUserDialog
                onCancel={handleCloseVerifyDeleteUserDialog}
                open={isVerifyDeleteUserDialogOpen}
                username={userToDelete || ""}
            />
            <SetPasswordDialog
                onCancel={handleClosePasswordChangeDialog}
                open={isPasswordChangeDialogOpen}
                username={userForPasswordChange || ""}
            />
            <VerifyActionDialog
                cancelText={translate("Cancel")}
                confirmText={translate("Send Password Reset Email")}
                message={
                    translate(
                        "You are about to send a password reset email to {{user}} {{username}}, would you like to continue",
                        { user: translate("User").toLowerCase(), username: userForPasswordReset },
                    ) + "?"
                }
                onCancel={handleClosePasswordResetDialog}
                onConfirm={handleSendPasswordResetEmail}
                open={isPasswordResetDialogOpen}
                title={translate("Reset Password")}
            />

            <DropdownMenu
                onOpenChange={(open) => {
                    if (!open) handleCloseMoreMenu();
                }}
                open={Boolean(moreMenuAnchor)}
            >
                <DropdownMenuContent align="end" anchor={moreMenuAnchor?.element} id="user-row-menu">
                    <DropdownMenuItem id="user-menu-reset-email" onClick={handleMenuResetEmail}>
                        <Mail />
                        {translate("Send Password Reset Email")}
                    </DropdownMenuItem>
                    <DropdownMenuItem id="user-menu-change-password" onClick={handleMenuChangePassword}>
                        <KeyRound />
                        {translate("Change Password")}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem id="user-menu-edit" onClick={handleMenuEdit}>
                        <Pencil />
                        {translate("Edit User")}
                    </DropdownMenuItem>
                    <DropdownMenuItem id="user-menu-delete" onClick={handleMenuDelete} variant="destructive">
                        <Trash2 />
                        {translate("Delete User")}
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            <div className="mb-4">
                <Button id="user-management-add" onClick={handleOpenNewUserDialog}>
                    {translate("Add a {{item}}", { item: translate("User").toLowerCase() })}
                </Button>
            </div>

            <DataTable
                columns={columns}
                emptyText={translate("No {{item}} found", { item: translate("User").toLowerCase() })}
                filter={filter}
                getRowId={(row) => row.username}
                id="user-management-table"
                loading={loading}
                onFilterChange={setFilter}
                onPaginationChange={(next) => setPagination(next)}
                onRowDoubleClick={(row) => handleOpenEditUserDialog(row.username)}
                onSortChange={setSort}
                pagination={{ ...pagination, total: filteredSortedUsers.length }}
                rowActions={rowActions}
                rowClassPrefix="user-row-"
                rows={rows}
                sort={sort}
            />
        </div>
    );
};

export default UserManagementView;
