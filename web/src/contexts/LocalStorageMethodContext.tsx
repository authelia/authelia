import React, { createContext, useCallback, useContext, useEffect, useState } from "react";

import { LocalStorageSecondFactorMethod } from "@constants/LocalStorage";
import { SecondFactorMethod } from "@models/Methods";
import { localStorageAvailable } from "@services/LocalStorage";
import { Method2FA, isMethod2FA, toMethod2FA, toSecondFactorMethod } from "@services/UserInfo";

export const LocalStorageMethodContext = createContext<ValueProps | null>(null);

export interface Props {
    children: React.ReactNode;
}

export interface ValueProps {
    localStorageMethod: SecondFactorMethod | undefined;
    setLocalStorageMethod: (value: SecondFactorMethod | undefined) => void;
    localStorageMethodAvailable: boolean;
}

export default function LocalStorageMethodContextProvider(props: Props) {
    const [localStorageMethod, setLocalStorageMethod] = useState<SecondFactorMethod>();
    const localStorageMethodAvailable = localStorageAvailable();

    const callback = useCallback((value: SecondFactorMethod | undefined) => {
        setLocalStorageMethod(value);

        if (!value) {
            window.localStorage.removeItem(LocalStorageSecondFactorMethod);
        } else {
            window.localStorage.setItem(LocalStorageSecondFactorMethod, toMethod2FA(value));
        }
    }, []);

    const listener = (ev: globalThis.StorageEvent): any => {
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
    };

    const refresh = useCallback(() => {
        const value = window.localStorage.getItem(LocalStorageSecondFactorMethod);

        if (value && isMethod2FA(value)) {
            if (isMethod2FA(value)) {
                return setLocalStorageMethod(toSecondFactorMethod(value as Method2FA));
            }

            return false;
        }
    }, []);

    useEffect(() => {
        if (!localStorageMethodAvailable) return;

        refresh();

        window.addEventListener("storage", listener);

        return () => {
            window.removeEventListener("storage", listener);
        };
    }, [localStorageMethodAvailable, refresh]);

    return (
        <LocalStorageMethodContext.Provider
            value={{
                localStorageMethod: localStorageMethod,
                setLocalStorageMethod: callback,
                localStorageMethodAvailable: localStorageMethodAvailable,
            }}
        >
            {props.children}
        </LocalStorageMethodContext.Provider>
    );
}

export function useLocalStorageMethodContext() {
    const context = useContext(LocalStorageMethodContext);
    if (!context) {
        throw new Error("useLocalStorageMethodContext must be used within a LocalStorageMethodContextProvider");
    }

    return context;
}
