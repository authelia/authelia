import { useCallback, useEffect, useMemo, useState } from "react";

import { Plus, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";

import { ColumnDef, DataTable, RowAction } from "@components/DataTable";
import { Button } from "@components/UI/Button";
import { useNotifications } from "@contexts/NotificationsContext";
import { useAllGroupsGET } from "@hooks/GroupManagement";
import NewGroupDialog from "@views/Settings/UserManagement/NewGroupDialog";
import VerifyDeleteGroupDialog from "@views/Settings/UserManagement/VerifyDeleteGroupDialog";

interface GroupRow {
    name: string;
}

const GroupManagementView = () => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification } = useNotifications();

    const [groups, fetchGroups, loading, fetchGroupsError] = useAllGroupsGET();

    const [groupToDelete, setGroupToDelete] = useState("");
    const [isNewGroupDialogOpen, setIsNewGroupDialogOpen] = useState(false);
    const [isVerifyDeleteGroupDialogOpen, setIsVerifyDeleteGroupDialogOpen] = useState(false);

    useEffect(() => {
        fetchGroups();
    }, [fetchGroups]);

    useEffect(() => {
        if (fetchGroupsError) {
            createErrorNotification(translate("There was an issue retrieving group info"));
        }
    }, [createErrorNotification, fetchGroupsError, translate]);

    const handleOpenNewGroupDialog = useCallback(() => {
        setIsNewGroupDialogOpen(true);
    }, []);

    const handleCloseNewGroupDialog = useCallback(() => {
        setIsNewGroupDialogOpen(false);
        fetchGroups();
    }, [fetchGroups]);

    const handleOpenVerifyDeleteGroupDialog = useCallback((groupName: string) => {
        setGroupToDelete(groupName);
        setIsVerifyDeleteGroupDialogOpen(true);
    }, []);

    const handleCloseVerifyDeleteGroupDialog = useCallback(() => {
        setIsVerifyDeleteGroupDialogOpen(false);
        fetchGroups();
    }, [fetchGroups]);

    const columns = useMemo<ColumnDef<GroupRow>[]>(
        () => [{ field: "name", header: translate("Group Name"), value: (row) => row.name }],
        [translate],
    );

    const rowActions = useMemo<RowAction<GroupRow>[]>(
        () => [
            {
                destructive: true,
                icon: <Trash2 />,
                id: () => "delete",
                label: translate("Delete this {{item}}", { item: translate("Group") }),
                onClick: (row) => handleOpenVerifyDeleteGroupDialog(row.name),
            },
        ],
        [handleOpenVerifyDeleteGroupDialog, translate],
    );

    const rows = useMemo<GroupRow[]>(() => (groups ?? []).map((name) => ({ name })), [groups]);

    return (
        <div className="flex h-[calc(100dvh-3.5rem)] flex-col overflow-hidden px-4 pt-4 pb-6 sm:h-[calc(100dvh-6.5rem)] sm:px-0 sm:pt-0">
            <h4 className="mb-4 shrink-0 text-2xl font-semibold">{translate("Group Management")}</h4>

            <NewGroupDialog onClose={handleCloseNewGroupDialog} open={isNewGroupDialogOpen} />
            <VerifyDeleteGroupDialog
                groupName={groupToDelete || ""}
                onCancel={handleCloseVerifyDeleteGroupDialog}
                open={isVerifyDeleteGroupDialogOpen}
            />

            <DataTable
                columns={columns}
                emptyText={translate("No {{item}} found", { item: translate("Group").toLowerCase() })}
                getRowId={(row) => row.name}
                id="group-management-table"
                initialSort={{ direction: "asc", field: "name" }}
                loading={loading}
                rowActions={rowActions}
                rowClassPrefix="group-row-"
                rows={rows}
                toolbarEnd={
                    <Button id="group-management-add" onClick={handleOpenNewGroupDialog}>
                        <Plus />
                        {translate("Add a {{item}}", { item: translate("Group").toLowerCase() })}
                    </Button>
                }
            />
        </div>
    );
};

export default GroupManagementView;
