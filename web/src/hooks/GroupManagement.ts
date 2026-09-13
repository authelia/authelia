import { useRemoteCall } from "@hooks/RemoteCall";
import { deleteGroup, getAllGroups, postNewGroup } from "@services/GroupManagement";

export function useAllGroupsGET() {
    return useRemoteCall(getAllGroups);
}

export function useGroupPOST() {
    return useRemoteCall(postNewGroup);
}

export function useGroupDELETE() {
    return useRemoteCall(deleteGroup);
}
