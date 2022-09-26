import queryString from "query-string";
import { useLocation } from "react-router-dom";

export interface Workflow {
    id?: string;
    name: string;
}

export function useWorkflow() {
    const location = useLocation();
    const queryParams = queryString.parse(location.search);
    if (!queryParams || !("workflow" in queryParams)) {
        return undefined;
    }

    let workflow = queryParams["workflow"] as string;

    return queryParams && "workflow_id" in queryParams
        ? ({ id: queryParams["workflow"] as string, name: workflow } as Workflow)
        : ({ name: workflow } as Workflow);
}
