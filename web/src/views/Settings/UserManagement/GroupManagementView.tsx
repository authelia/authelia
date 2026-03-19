import { useCallback, useEffect, useMemo, useState } from "react";

import DeleteIcon from "@mui/icons-material/DeleteOutlined";
import { Button, Stack } from "@mui/material";
import { DataGrid, GridActionsCellItem, GridColDef, GridRowParams, GridRowsProp } from "@mui/x-data-grid";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { useAllGroupsGET } from "@hooks/GroupManagement";
import NewGroupDialog from "@views/Settings/UserManagement/NewGroupDialog";
import VerifyDeleteGroupDialog from "@views/Settings/UserManagement/VerifyDeleteGroupDialog";

const GroupManagementView = () => {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification } = useNotifications();

    const [groups, fetchGroups, , fetchGroupsError] = useAllGroupsGET();
    const [groupToDelete, setGroupToDelete] = useState("");
    const [isNewGroupDialogOpen, setIsNewGroupDialogOpen] = useState(false);
    const [isVerifyDeleteGroupDialogOpen, setIsVerifyDeleteGroupDialogOpen] = useState(false);

    const handleResetState = useCallback(() => {
        setIsVerifyDeleteGroupDialogOpen(false);
        setIsNewGroupDialogOpen(false);
    }, []);

    const handleOpenNewGroupDialog = useCallback(() => {
        setIsNewGroupDialogOpen(true);
    }, []);

    const handleCloseNewGroupDialog = useCallback(() => {
        setIsNewGroupDialogOpen(false);
        handleResetState();
        fetchGroups();
    }, [handleResetState, fetchGroups]);

    const handleOpenVerifyDeleteGroupDialog = useCallback(
        (groupName: string) => {
            setGroupToDelete(groupName);
            setIsVerifyDeleteGroupDialogOpen(true);
        },
        [],
    );

    const handleCloseVerifyDeleteGroupDialog = useCallback(() => {
        setIsVerifyDeleteGroupDialogOpen(false);
        handleResetState();
        fetchGroups();
    }, [handleResetState, fetchGroups]);

    useEffect(() => {
        fetchGroups();
    }, [fetchGroups]);

    useEffect(() => {
        if (fetchGroupsError) {
            createErrorNotification(translate("There was an issue retrieving group info"));
        }
    }, [fetchGroupsError, createErrorNotification, translate]);

    const rows: GridRowsProp = useMemo(() => {
        if (!groups) {
            return [];
        }

        if (!Array.isArray(groups)) {
            createErrorNotification("Error fetching Group Info");
            return [];
        }

        return groups.map((group: string, index: number) => ({
            id: index,
            name: group,
        }));
    }, [groups, createErrorNotification]);

    const columns: GridColDef[] = [
        { field: "name", flex: 1, headerName: "Group Name" },
        {
            cellClassName: "actions",
            field: "actions",
            getActions: (params: GridRowParams) => {
                return [
                    <GridActionsCellItem
                        key="delete"
                        icon={<DeleteIcon />}
                        label="Delete"
                        onClick={() => handleOpenVerifyDeleteGroupDialog(params.row.name)}
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
            <NewGroupDialog open={isNewGroupDialogOpen} onClose={handleCloseNewGroupDialog} />
            <VerifyDeleteGroupDialog
                groupName={groupToDelete || ""}
                open={isVerifyDeleteGroupDialogOpen}
                onCancel={handleCloseVerifyDeleteGroupDialog}
            />
            <Stack direction={"row"} spacing={1} sx={{ mb: 1 }}>
                <Button size="medium" variant="contained" onClick={handleOpenNewGroupDialog}>
                    {translate("Add a {{item}}", { item: "group" })}
                </Button>
            </Stack>
            <div style={{ height: 400, width: "100%" }}>
                <DataGrid
                    rows={rows}
                    columns={columns}
                    checkboxSelection={false}
                    initialState={{
                        sorting: {
                            sortModel: [{ field: "name", sort: "asc" }],
                        },
                        pagination: {
                            paginationModel: { pageSize: 25, page: 0 },
                        },
                    }}
                />
            </div>
        </>
    );
};

export default GroupManagementView;
