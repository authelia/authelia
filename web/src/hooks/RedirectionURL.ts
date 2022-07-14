import queryString from "query-string";
import { useLocation } from "react-router-dom";

export function useRedirectionURL() {
    const location = useLocation();

    const queryParams = queryString.parse(location.search);

    return queryParams && "rd" in queryParams ? (queryParams["rd"] as string) : undefined;
}
