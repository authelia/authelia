import { useCallback } from "react";

import { useNavigate, useSearchParams } from "react-router-dom";

import { QueryParamFlow, QueryParamFlowID, QueryParamRedirection, QueryParamSubFlow } from "@constants/constants";

export function useRouterNavigate() {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();

    return useCallback(
        (
            pathname: string,
            preserveSearchParams: boolean = true,
            preserveFlow: boolean = true,
            preserveRedirection: boolean = true,
            searchParamsOverride: URLSearchParams | undefined = undefined,
        ) => {
            if (searchParamsOverride && URLSearchParamsHasValues(searchParamsOverride)) {
                navigate({ pathname: pathname, search: `?${searchParamsOverride.toString()}` });
            } else if (URLSearchParamsHasValues(searchParams)) {
                if (preserveSearchParams) {
                    navigate({ pathname: pathname, search: `?${searchParams.toString()}` });
                } else if (preserveFlow || preserveRedirection) {
                    const redirection = searchParams?.get(QueryParamRedirection);
                    const flow = searchParams?.get(QueryParamFlow);
                    const flowID = searchParams?.get(QueryParamFlowID);
                    const subflow = searchParams?.get(QueryParamSubFlow);

                    const params = new URLSearchParams();

                    if (preserveRedirection && redirection) params.set(QueryParamRedirection, redirection);

                    if (preserveFlow) {
                        if (flow) params.set(QueryParamFlow, flow);
                        if (flowID) params.set(QueryParamFlowID, flowID);
                        if (subflow) params.set(QueryParamSubFlow, subflow);
                    }

                    navigate({ pathname: pathname, search: `?${params.toString()}` });
                } else {
                    navigate({ pathname: pathname });
                }
            } else {
                navigate({ pathname: pathname });
            }
        },
        [navigate, searchParams],
    );
}

function URLSearchParamsHasValues(params?: URLSearchParams) {
    return params ? !params.entries().next().done : false;
}
