import { useSearchParams } from "react-router-dom";

import { Flow, FlowID, RedirectionURL, SubFlow } from "@constants/SearchParams";

export function useFlow(): { id?: string; flow?: string; subflow?: string } {
    const [query] = useSearchParams();

    const id = query.get(FlowID);
    const flow = query.get(Flow);
    const subflow = query.get(SubFlow);

    return {
        flow: flow ?? undefined,
        id: id ?? undefined,
        subflow: subflow ?? undefined,
    };
}

export function useFlowPresent(): boolean {
    const [query] = useSearchParams();

    const flow = query.get(Flow);
    const redirection = query.get(RedirectionURL);

    return flow !== null || redirection !== null;
}
