import { KeyboardEvent } from "react";

const safe = /^[0-9!@#$%^&*)(+=[{\]};:'",<.>/?\\|`~_-]$/i;

export function IsCapsLockModified(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key.length !== 1) return null;
    if (event.ctrlKey || event.altKey || event.metaKey) return null;
    if (event.key === " ") return null;
    if (safe.test(event.key)) return null;

    return event.getModifierState("CapsLock");
}
