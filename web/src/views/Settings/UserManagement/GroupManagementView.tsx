import { useCallback, useEffect, useMemo, useState } from "react";

import { Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";

import { ColumnDef, DataTable, PaginationState, RowAction, SortState } from "@components/DataTable";
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

    const [sort, setSort] = useState<SortState>({ direction: "asc", field: "name" });
    const [pagination, setPagination] = useState<Omit<PaginationState, "total">>({ page: 1, pageSize: 25 });
    const [filter, setFilter] = useState("");

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
                icon: <Trash2 />,
                id: () => "delete",
                label: translate("Delete this {{item}}", { item: translate("Group") }),
                onClick: (row) => handleOpenVerifyDeleteGroupDialog(row.name),
            },
        ],
        [handleOpenVerifyDeleteGroupDialog, translate],
    );

    const filteredSortedGroups = useMemo(() => {
        const all = (groups ?? []).map((name): GroupRow => ({ name }));

        const filtered = all.filter((row) => !filter || row.name.toLowerCase().includes(filter.toLowerCase()));

        const sorted = [...filtered].sort((a, b) => {
            const result = a.name.localeCompare(b.name);

            return sort.direction === "asc" ? result : -result;
        });

        return sorted;
    }, [filter, groups, sort.direction]);

    const rows = useMemo(
        () =>
            filteredSortedGroups.slice(
                (pagination.page - 1) * pagination.pageSize,
                pagination.page * pagination.pageSize,
            ),
        [filteredSortedGroups, pagination.page, pagination.pageSize],
    );

    return (
        <div className="flex flex-col">
            <h4 className="mb-4 text-2xl font-semibold">{translate("Group Management")}</h4>

            <NewGroupDialog onClose={handleCloseNewGroupDialog} open={isNewGroupDialogOpen} />
            <VerifyDeleteGroupDialog
                groupName={groupToDelete || ""}
                onCancel={handleCloseVerifyDeleteGroupDialog}
                open={isVerifyDeleteGroupDialogOpen}
            />

            <div className="mb-4">
                <Button id="group-management-add" onClick={handleOpenNewGroupDialog}>
                    {translate("Add a {{item}}", { item: translate("Group").toLowerCase() })}
                </Button>
            </div>

            <DataTable
                columns={columns}
                emptyText={translate("No {{item}} found", { item: translate("Group").toLowerCase() })}
                filter={filter}
                getRowId={(row) => row.name}
                id="group-management-table"
                loading={loading}
                onFilterChange={setFilter}
                onPaginationChange={(next) => setPagination(next)}
                onSortChange={setSort}
                pagination={{ ...pagination, total: filteredSortedGroups.length }}
                rowActions={rowActions}
                rowClassPrefix="group-row-"
                rows={rows}
                sort={sort}
            />
        </div>
    );
};

export default GroupManagementView;
