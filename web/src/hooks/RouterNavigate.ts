import { useCallback } from "react";

import { useNavigate, useSearchParams } from "react-router-dom";

import { Flow, FlowID, RedirectionURL, SubFlow } from "@constants/SearchParams";

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
                    const params = new URLSearchParams();

                    if (preserveRedirection) {
                        const redirection = searchParams?.get(RedirectionURL);

                        if (redirection) params.set(RedirectionURL, redirection);
                    }

                    if (preserveFlow) {
                        const flow = searchParams?.get(Flow);
                        const subflow = searchParams?.get(SubFlow);
                        const flowID = searchParams?.get(FlowID);

                        if (flow) params.set(Flow, flow);
                        if (subflow) params.set(SubFlow, subflow);
                        if (flowID) params.set(FlowID, flowID);
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
