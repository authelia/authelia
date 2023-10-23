import { useSearchParams } from "react-router-dom";

export function useID(): string | undefined {
    const [searchParams] = useSearchParams();

    const id = searchParams.get("id");

    return id === null ? undefined : id;
}
