import { useSearchParams } from "react-router-dom";

import { Flow, FlowID, SubFlow } from "@constants/SearchParams";

export function useFlow(): { id?: string; flow?: string; subflow?: string } {
    const [searchParams] = useSearchParams();

    const id = searchParams.get(FlowID);
    const flow = searchParams.get(Flow);
    const subflow = searchParams.get(SubFlow);

    return {
        id: id === null ? undefined : id,
        flow: flow === null ? undefined : flow,
        subflow: subflow === null ? undefined : subflow,
    };
}
