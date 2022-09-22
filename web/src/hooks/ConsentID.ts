import queryString from "query-string";
import { useLocation } from "react-router-dom";

export function useConsentID() {
    const location = useLocation();
    const queryParams = queryString.parse(location.search);
    return queryParams && "consent_id" in queryParams ? (queryParams["consent_id"] as string) : undefined;
}
