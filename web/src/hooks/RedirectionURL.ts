import queryString from "query-string";
import { useLocation, useSearchParams } from "react-router-dom";

import { RedirectionURL, RequestMethod } from "@constants/SearchParams";

export function useRedirectionURL() {
    const location = useLocation();

    const queryParams = queryString.parse(location.search);

    return queryParams && RedirectionURL in queryParams ? (queryParams[RedirectionURL] as string) : undefined;
}

export function useRDRM(): [rd: string | null, rm: string | null] {
    const [searchParams] = useSearchParams();

    return [searchParams.get(RedirectionURL), searchParams.get(RequestMethod)];
}
