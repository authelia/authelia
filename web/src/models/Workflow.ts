export enum WorkflowType {
    None = 0,
    OpenIDConnect = 1,
}

export type WorkflowName = "none" | "openid_connect";

export interface Workflow {
    id: string;
    type: WorkflowType;
}

export function toWorkflowType(workflow: WorkflowName): WorkflowType {
    switch (workflow) {
        case "none":
            return WorkflowType.None;
        case "openid_connect":
            return WorkflowType.OpenIDConnect;
    }
}

export function toWorkflowName(workflow: WorkflowType): WorkflowName {
    switch (workflow) {
        case WorkflowType.None:
            return "none";
        case WorkflowType.OpenIDConnect:
            return "openid_connect";
    }
}

export function toWorkflowPath(path: string, workflow: Workflow): string {
    if (workflow.type !== WorkflowType.None) {
        path += `?workflow=${toWorkflowName(workflow.type)}`;

        if (workflow.id !== "") {
            path += `&workflow_id=${workflow.id}`;
        }
    }

    return path;
}
