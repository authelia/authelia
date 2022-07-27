import queryString from "query-string";
import { useLocation } from "react-router-dom";

export function useWorkflow() {
    const location = useLocation();
    const queryParams = queryString.parse(location.search);
    return queryParams && "workflow" in queryParams ? (queryParams["workflow"] as string) : undefined;
}
