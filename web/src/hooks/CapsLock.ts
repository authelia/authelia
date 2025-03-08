import React, { useCallback } from "react";

export const useCheckCapsLock = (setCapsLockNotify: React.Dispatch<React.SetStateAction<boolean>>) => {
    return useCallback(
        (event: React.KeyboardEvent<HTMLDivElement>) => {
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
