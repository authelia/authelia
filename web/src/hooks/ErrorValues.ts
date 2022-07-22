import queryString from "query-string";
import { useLocation } from "react-router-dom";

interface ErrorValues {
    title?: string;
    description?: string;
    code?: string;
    message?: string;
    url?: string;
}

export function useErrorValues() {
    const location = useLocation();
    const queryParams = queryString.parse(location.search);

    return {
        title: queryParams && "title" in queryParams ? (queryParams["title"] as string) : undefined,
        description: queryParams && "description" in queryParams ? (queryParams["description"] as string) : undefined,
        code: queryParams && "code" in queryParams ? (queryParams["code"] as string) : undefined,
        message: queryParams && "message" in queryParams ? (queryParams["message"] as string) : undefined,
        url: queryParams && "url" in queryParams ? (queryParams["url"] as string) : undefined,
    } as ErrorValues;
}
