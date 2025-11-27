import { useCallback } from "react";

import { useNavigate, useSearchParams } from "react-router-dom";

import { Flow, FlowID, RedirectionURL, SubFlow, UserCode } from "@constants/SearchParams";

export function useRouterNavigate() {
    const navigate = useNavigate();
    const [query] = useSearchParams();

    return useCallback(
        (
            pathname: string,
            preserveSearchParams: boolean = true,
            preserveFlow: boolean = true,
            preserveRedirection: boolean = true,
            searchParamsOverride: undefined | URLSearchParams = undefined,
        ) => {
            if (searchParamsOverride && URLSearchParamsHasValues(searchParamsOverride)) {
                navigate({ pathname: pathname, search: `?${searchParamsOverride.toString()}` });
            } else if (URLSearchParamsHasValues(query)) {
                if (preserveSearchParams) {
                    navigate({ pathname: pathname, search: `?${query.toString()}` });
                } else if (preserveFlow || preserveRedirection) {
                    const params = new URLSearchParams();

                    if (preserveRedirection) {
                        const redirection = query?.get(RedirectionURL);

                        if (redirection) params.set(RedirectionURL, redirection);
                    }

                    if (preserveFlow) {
                        const flow = query?.get(Flow);
                        const subflow = query?.get(SubFlow);
                        const flowID = query?.get(FlowID);
                        const userCode = query?.get(UserCode);

                        if (flow) params.set(Flow, flow);
                        if (subflow) params.set(SubFlow, subflow);
                        if (flowID) params.set(FlowID, flowID);
                        if (userCode) params.set(UserCode, userCode);
                    }

                    navigate({ pathname: pathname, search: `?${params.toString()}` });
                } else {
                    navigate({ pathname: pathname });
                }
            } else {
                navigate({ pathname: pathname });
            }
        },
        [navigate, query],
    );
}

function URLSearchParamsHasValues(params?: URLSearchParams) {
    return params ? !params.entries().next().done : false;
}
