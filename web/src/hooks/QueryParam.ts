import { useSearchParams } from "react-router-dom";

export function useQueryParam(queryParam: string) {
    const [query] = useSearchParams();
    const value = query.get(queryParam);
    return value !== "" ? (value as string) : undefined;
}
