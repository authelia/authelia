import { useRemoteCall } from "@hooks/RemoteCall";
import { getAllGroups } from "@services/GroupManagement";

export function useAllGroupsGET() {
    return useRemoteCall(getAllGroups);
}
