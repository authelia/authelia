import queryString from "query-string";
import { useLocation } from "react-router-dom";

export function useRedirectionURL() {
    const location = useLocation();

    console.log("Location Search is ", location.search);

    const queryParams = queryString.parse(location.search);

    console.log("Query Params is ", queryParams);

    const result = queryParams && "rd" in queryParams ? (queryParams["rd"] as string) : undefined;

    console.log("Result is ", result);

    return result;
}
