import { ReactNode, createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

import { LocalStorageSecondFactorMethod } from "@constants/LocalStorage";
import { SecondFactorMethod } from "@models/Methods";
import { localStorageAvailable } from "@services/LocalStorage";
import { Method2FA, isMethod2FA, toMethod2FA, toSecondFactorMethod } from "@services/UserInfo";

export const LocalStorageMethodContext = createContext<null | ValueProps>(null);

export interface Props {
    readonly children: ReactNode;
}

export interface ValueProps {
    localStorageMethod: SecondFactorMethod | undefined;
    setLocalStorageMethod: (_value: SecondFactorMethod | undefined) => void;
    localStorageMethodAvailable: boolean;
}

export default function LocalStorageMethodContextProvider(props: Props) {
    const localStorageMethodAvailable = localStorageAvailable();

    const [localStorageMethod, setLocalStorageMethod] = useState<SecondFactorMethod | undefined>(() => {
        if (!localStorageMethodAvailable) return undefined;

        const value = globalThis.localStorage.getItem(LocalStorageSecondFactorMethod);
        if (value && isMethod2FA(value)) {
            return toSecondFactorMethod(value as Method2FA);
        }
        return undefined;
    });

    const callback = useCallback((value: SecondFactorMethod | undefined) => {
        setLocalStorageMethod(value);

        if (value) {
            globalThis.localStorage.setItem(LocalStorageSecondFactorMethod, toMethod2FA(value));
        } else {
            globalThis.localStorage.removeItem(LocalStorageSecondFactorMethod);
        }
    }, []);

    const listener = useCallback((ev: globalThis.StorageEvent): any => {
        if (ev.key !== LocalStorageSecondFactorMethod) {
            return;
        }

        if (ev.newValue) {
            if (isMethod2FA(ev.newValue)) {
                setLocalStorageMethod(toSecondFactorMethod(ev.newValue as Method2FA));
            }
        } else {
            setLocalStorageMethod(undefined);
        }
    }, []);

    useEffect(() => {
        if (!localStorageMethodAvailable) return;

        globalThis.addEventListener("storage", listener);

        return () => {
            globalThis.removeEventListener("storage", listener);
        };
    }, [localStorageMethodAvailable, listener]);

    const value = useMemo(
        () => ({
            localStorageMethod: localStorageMethod,
            localStorageMethodAvailable: localStorageMethodAvailable,
            setLocalStorageMethod: callback,
        }),
        [localStorageMethod, callback, localStorageMethodAvailable],
    );

    return <LocalStorageMethodContext.Provider value={value}>{props.children}</LocalStorageMethodContext.Provider>;
}

export function useLocalStorageMethodContext() {
    const context = useContext(LocalStorageMethodContext);
    if (!context) {
        throw new Error("useLocalStorageMethodContext must be used within a LocalStorageMethodContextProvider");
    }

    return context;
}
