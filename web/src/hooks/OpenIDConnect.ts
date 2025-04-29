import { useSearchParams } from "react-router-dom";

export function useUserCode() {
    const [searchParams] = useSearchParams();

    const userCode = searchParams.get(QueryParamUserCode);

    return userCode === null ? undefined : userCode;
}

export const QueryParamUserCode = "user_code";
