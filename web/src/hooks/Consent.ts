import { useRemoteCall } from "@hooks/RemoteCall";
import { Workflow } from "@models/Workflow";
import { getRequestedScopesWorkflow } from "@services/Consent";

export function useRequestedScopes(workflow: Workflow) {
    return useRemoteCall(getRequestedScopesWorkflow, [workflow]);
}
