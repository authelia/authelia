import { useSearchParams } from "react-router-dom";

export function useUserCode() {
    const [searchParams] = useSearchParams();

    return searchParams.get(QueryParamUserCode);
}

export const QueryParamUserCode = "user_code";
