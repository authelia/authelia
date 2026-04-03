import { Dispatch, KeyboardEvent, SetStateAction, useCallback } from "react";

export const useCheckCapsLock = (setCapsLockNotify: Dispatch<SetStateAction<boolean>>) => {
    return useCallback(
        (event: KeyboardEvent<HTMLDivElement>) => {
            if (event.getModifierState("CapsLock")) {
                setCapsLockNotify(true);
            } else {
                setCapsLockNotify(false);
            }
        },
        [setCapsLockNotify],
    );
};

export default useCheckCapsLock;
