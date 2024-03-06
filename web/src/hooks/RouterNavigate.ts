import { useCallback } from "react";

import { useNavigate, useSearchParams } from "react-router-dom";

export function useRouterNavigate() {
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();

    return useCallback(
        (
            pathname: string,
            preserveSearchParams: boolean = true,
            searchParamsOverride: URLSearchParams | undefined = undefined,
        ) => {
            if (searchParamsOverride && URLSearchParamsHasValues(searchParamsOverride)) {
                navigate({ pathname: pathname, search: `?${searchParamsOverride.toString()}` });
            } else if (preserveSearchParams && URLSearchParamsHasValues(searchParams)) {
                navigate({ pathname: pathname, search: `?${searchParams.toString()}` });
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
