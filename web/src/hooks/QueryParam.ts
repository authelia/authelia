import { useSearchParams } from "react-router-dom";

export function useQueryParam(queryParam: string) {
    const [searchParams] = useSearchParams();
    const value = searchParams.get(queryParam);
    return value !== "" && value !== null ? (value as string) : undefined;
}
