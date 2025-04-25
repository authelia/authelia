import { useSearchParams } from "react-router-dom";

import { QueryParamFlow, QueryParamFlowID, QueryParamSubFlow } from "@constants/constants";

export function useFlow(): { id?: string; flow?: string; subflow?: string } {
    const [searchParams] = useSearchParams();

    const id = searchParams.get(QueryParamFlowID);
    const flow = searchParams.get(QueryParamFlow);
    const subflow = searchParams.get(QueryParamSubFlow);

    return {
        id: id === null ? undefined : id,
        flow: flow === null ? undefined : flow,
        subflow: subflow === null ? undefined : subflow,
    };
}
