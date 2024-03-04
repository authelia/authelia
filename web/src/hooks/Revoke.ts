import { useSearchParams } from "react-router-dom";

export function useID(): string | undefined {
    const [searchParams] = useSearchParams();

    const id = searchParams.get("id");

    return id === null ? undefined : id;
}

export function useToken(): string | undefined {
    const [searchParams] = useSearchParams();

    const token = searchParams.get("token");

    return token === null ? undefined : token;
}
