import { useSearchParams } from "react-router-dom";

import { Workflow, WorkflowIdentifier } from "@constants/SearchParams";

export function useWorkflow(): [workflow: string | null, workflowID: string | null] {
    const [searchParams] = useSearchParams();

    return [searchParams.get(Workflow), searchParams.get(WorkflowIdentifier)];
}
