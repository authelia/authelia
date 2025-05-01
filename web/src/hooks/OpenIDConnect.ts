import { useSearchParams } from "react-router-dom";

import { UserCode } from "@constants/SearchParams";

export function useUserCode() {
    const [searchParams] = useSearchParams();

    const userCode = searchParams.get(UserCode);

    return userCode === null ? undefined : userCode;
}
