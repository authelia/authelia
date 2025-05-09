import { useSearchParams } from "react-router-dom";

import { UserCode } from "@constants/SearchParams";

export function useUserCode() {
    const [query] = useSearchParams();

    const userCode = query.get(UserCode);

    return userCode === null ? undefined : userCode;
}
