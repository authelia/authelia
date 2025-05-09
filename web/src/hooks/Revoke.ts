import { useSearchParams } from "react-router-dom";

import { Identifier, IdentityToken } from "@constants/SearchParams";

export function useID(): string | undefined {
    const [query] = useSearchParams();

    const id = query.get(Identifier);

    return id === null ? undefined : id;
}

export function useToken(): string | undefined {
    const [query] = useSearchParams();

    const token = query.get(IdentityToken);

    return token === null ? undefined : token;
}
