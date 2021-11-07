import queryString from "query-string";
import { useLocation } from "react-router-dom";

export function useRequestMethod() {
    const location = useLocation();
    const queryParams = queryString.parse(location.search);
    return queryParams && "rm" in queryParams ? (queryParams["rm"] as string) : undefined;
}
