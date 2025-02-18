import { useLocation } from "react-router-dom";

export function useUserCode() {
    const { hash } = useLocation();

    let raw = hash;

    if (raw.startsWith("#")) {
        raw = raw.slice(1);
    }

    const parameters = new URLSearchParams(raw);

    return parameters.get("user_code");
}
