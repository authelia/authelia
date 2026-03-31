import { AdminGroupRestPath } from "@services/Api";
import { DeleteWithOptionalResponse, Get, PostWithOptionalResponse } from "@services/Client";

export interface NewGroupRequest {
    name: string;
}

export async function getAllGroups(): Promise<string[]> {
    return await Get<string[]>(AdminGroupRestPath);
}

export async function postNewGroup(groupData: NewGroupRequest) {
    return PostWithOptionalResponse(AdminGroupRestPath, groupData);
}

export async function deleteGroup(groupName: string) {
    return DeleteWithOptionalResponse(`${AdminGroupRestPath}/${groupName}`);
}
